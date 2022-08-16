//
//  Protocol.swift
//  ultrablue
//
//  Created by loic buckwell on 22/07/2022.
//

import Foundation
import SwiftCBOR
import Gomobile
import CoreData

extension CaseIterable where Self: Equatable {
    private var allCases: AllCases { Self.allCases }
    var next: Self {
        let index = allCases.index(after: allCases.firstIndex(of: self)!)
        guard index != allCases.endIndex else { return allCases.first! }
        return allCases[index]
    }
}

enum ProtoState: CaseIterable {
    case enroll,
         enroll_handle,
         auth_read,
         auth_write,
         ak_read,
         make_credential,
         activate_credential_read,
         activate_credential_handle,
         anti_replay_nonce,
         attestation_data_read,
         quotes_signature_verify,
         event_log_replay,
         pcrs_read,
         pcr_policy_apply,
         attestation_response,
         end
    
    func next() -> ProtoState {
        return self.next
    }
}

class UltrablueProtocol {
    private var logger: Logger
    private var currentState: ProtoState
    private var bleManager: BLEManager
    private var onSuccess: () -> Void
    
    private var context: NSManagedObjectContext
    private var device: Device?
    private var enroll: Bool
    private var enrollData: EnrollDataModel? = nil
    private var authNonce: Data? = nil
    private var encodedak: Data? = nil
    private var encodedpp: Data? = nil
    private var credentialActivationSecret: Data? = nil
    private var antiReplayNonce: Data? = nil
    private var encodedPCRs: Data? = nil
    private var secret: Data = Data()
    private var attestationResponse: AttestationResponse? = nil
    
    init(device: Device?, context: NSManagedObjectContext, bleManager: BLEManager, logger: Logger, onSuccess: @escaping () -> Void) {
        self.context = context
        self.device = device
        self.enroll = device == nil
        self.currentState = enroll ? .enroll : .auth_read
        self.bleManager = bleManager
        self.logger = logger
        self.onSuccess = onSuccess
        if enroll == false {
            enrollData = EnrollDataModel(Cert: device!.cert!, N: device!.n!, E: UInt(device!.e), PCRExtend: device!.secret != nil)
        }
        bleManager.setOnMsgReadCallback(callback: self.onMsgRead)
        bleManager.setOnMsgWriteCallback(callback: self.onMsgWrite)
        startProtocol()
    }
    
    private func startProtocol() {
        resumeProtocol()
    }
    
    private func resumeProtocol(with msg: Data? = nil, at state: ProtoState? = nil) {
        if let s = state {
            currentState = s
        }
        switch currentState {
        case .enroll:
            readMsg(tag: "EK pub and certificate")
        case .enroll_handle:
            do {
                let decoder = CodableCBORDecoder()
                enrollData = try decoder.decode(EnrollDataModel.self, from: msg!)
            }
            catch {
                print("An error occured while decoding CBOR: \(error)")
            }
            resumeProtocol(at: .auth_read)
        case .auth_read:
            readMsg(tag: "encrypted auth nonce")
        case .auth_write:
            writeMsg(tag: "decrypted auth nonce", msg: msg!)
        case .ak_read:
            readMsg(tag: "attestation key")
        case .make_credential:
            logger.push(log: Log("Generating activation credential"))
            encodedak = msg!
            let credentialBlob = Gomobile.GomobileMakeCredential(enrollData!.N, Int(enrollData!.E), encodedak, nil)
            guard let cb = credentialBlob else {
                logger.completeLast(success: false)
                return
            }
            credentialActivationSecret = cb.secret
            let encryptedCredential = EncryptedCredentialModel(Credential: cb.cred!, Secret: cb.credSecret!)
            logger.completeLast(success: true)
            do {
                let encoder = CodableCBOREncoder()
                let encodedEncryptedCredential = try encoder.encode(encryptedCredential)
                writeMsg(tag: "encrypted activation credential", msg: encodedEncryptedCredential)
            } catch {
                print("An error occured while encoding CBOR: \(error)")
            }
        case .activate_credential_read:
            readMsg(tag: "decrypted activation credential")
        case .activate_credential_handle:
            do {
                let decoder = CodableCBORDecoder()
                let s = try decoder.decode(ByteString.self, from: msg!)
                logger.push(log: Log("Verifying credential"))
                guard credentialActivationSecret!.elementsEqual(s.Bytes) else {
                    logger.completeLast(success: false)
                    return
                }
                logger.completeLast(success: true)
                resumeProtocol(at: .anti_replay_nonce)
            }
            catch {
                print("An error occured while decoding CBOR: \(error)")
            }
        case .anti_replay_nonce:
            logger.push(log: Log("Generating anti-replay nonce"))
            var bytes = [Int8](repeating: 0, count: 16)
            guard SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes) == errSecSuccess else {
                logger.completeLast(success: false)
                return
            }
            antiReplayNonce = Data(bytes: &bytes, count: 16)
            logger.completeLast(success: true)
            do {
                let encoder = CodableCBOREncoder()
                let arn = ByteString(Bytes: antiReplayNonce!)
                let encodedAntiReplayNonce = try encoder.encode(arn)
                writeMsg(tag: "anti-replay nonce", msg: encodedAntiReplayNonce)
            } catch {
                print("An error occured while encoding CBOR: \(error)")
            }
        case .attestation_data_read:
            readMsg(tag: "attestation data")
        case .quotes_signature_verify:
            logger.push(log: Log("Checking quotes signature"))
            encodedpp = msg
            guard Gomobile.GomobileCheckQuotesSignature(encodedak, encodedpp, antiReplayNonce, nil) else {
                logger.completeLast(success: false)
                return
            }
            logger.completeLast(success: true)
            resumeProtocol(at: .event_log_replay)
        case .event_log_replay:
            logger.push(log: Log("Replaying event log"))
            guard Gomobile.GomobileReplayEventLog(encodedpp, nil) else {
                logger.completeLast(success: false)
                return
            }
            logger.completeLast(success: true)
            resumeProtocol(at: .pcrs_read)
        case .pcrs_read:
            logger.push(log: Log("Extract PCRs from attestation data"))
            let wrappedPCRs = Gomobile.GomobileGetPCRs(encodedpp, nil)
            guard let wp = wrappedPCRs else {
                logger.completeLast(success: false)
                return
            }
            encodedPCRs = wp.data
            logger.completeLast(success: true)
            if enroll {
                if enrollData?.PCRExtend == true {
                    logger.push(log: Log("Generating new attester secret"))
                    guard let s = generateSecret() else {
                        logger.completeLast(success: false)
                        return
                    }
                    logger.completeLast(success: true)
                    secret = s
                }
                attestationResponse = AttestationResponse(err: false, Secret: secret)
                resumeProtocol(at: .attestation_response)
            } else {
                resumeProtocol(at: .pcr_policy_apply)
            }
        case .pcr_policy_apply:
                logger.push(log: Log("Applying PCR policy"))
                // TODO: Checking each stored PCR against received ones
                logger.completeLast(success: false)
        case .attestation_response:
            guard let ar = attestationResponse else {
                print("attestation response must be set at this stage")
                return
            }
            do {
                let encoder = CodableCBOREncoder()
                let encoded = try encoder.encode(ar)
                writeMsg(tag: "attestation response", msg: encoded)
            }
            catch {
                print("An error occured while encoding CBOR: \(error)")
                return
            }
        case .end:
            saveNewDevice()
            onSuccess()
        }
    }
    
    private func saveNewDevice() {
        logger.push(log: Log("Saving new attester entry"))
        device = Device(context: context)
        guard let d = device else {
            logger.completeLast(success: false)
            return
        }
        d.id = UUID()
        d.creation_time = Date.now
        d.last_attestation_time = d.creation_time
        d.name = Name.generate()
        d.pcrpolicy = Policy(.strict).value
        d.e = Int64(enrollData!.E)
        d.n = enrollData!.N
        d.cert = enrollData!.Cert
        d.secret = secret
        d.addr = "\(bleManager.getAttesterUUID()!)"
        do {
            try context.save()
        } catch {
            logger.completeLast(success: false)
            return
        }
        logger.completeLast(success: true)
    }
    
    private func generateSecret() -> Data? {
        var bytes = [Int8](repeating: 0, count: 16)
        guard SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes) == errSecSuccess else {
            return nil
        }
        return Data(bytes: &bytes, count: 16)
    }
    
    private func readMsg(tag: String) {
        bleManager.setMessageTag(tag)
        bleManager.recvMsg()
    }
    
    private func writeMsg(tag: String, msg: Data) {
        bleManager.setMessageTag(tag)
        bleManager.sendMsg(msg: msg)
    }
    
    private func onMsgRead(msg: Data) {
        currentState = currentState.next()
        resumeProtocol(with: msg)
    }
    
    private func onMsgWrite() {
        currentState = currentState.next()
        resumeProtocol()
    }
}
