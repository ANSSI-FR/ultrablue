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
import CryptoKit

struct BootStateDiff {
    let PCRs: [IdentifiablePCR]
    let eventLogDiff: String
}

extension CaseIterable where Self: Equatable {
    private var allCases: AllCases { Self.allCases }
    var next: Self {
        let index = allCases.index(after: allCases.firstIndex(of: self)!)
        guard index != allCases.endIndex else { return allCases.first! }
        return allCases[index]
    }
}

enum ProtoState: CaseIterable {
    case uuid,
         start_encrypted_session,
         auth_read,
         auth_write,
         after_auth,
         enroll,
         enroll_handle,
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
         user_decision,
         attestation_response,
         end
    
    func next() -> ProtoState {
        return self.next
    }
}

class UltrablueProtocol {
    private var logger: Logger
    private var currentState: ProtoState = .uuid
    private var bleManager: BLEManager
    private var onCompletion: (Bool) -> Void
    private var onPCRChanged: (BootStateDiff) -> Void

    private var cipher: EncryptedSession? = nil

    private var context: NSManagedObjectContext
    private var enrollData: EnrollData? = nil
    private var device: Device? = nil
    private var enroll: Bool
    private var ek: EkModel? = nil
    private var authNonce: Data? = nil
    private var encodedak: Data? = nil
    private var encodedpp: Data? = nil
    private var credentialActivationSecret: Data? = nil
    private var antiReplayNonce: Data? = nil
    private var encodedPCRs: Data? = nil
    private var eventlog: String? = nil
    private var secret: Data = Data()
    private var attestationResponse: AttestationResponse? = nil
    
    init(device: Device, context: NSManagedObjectContext, bleManager: BLEManager, logger: Logger, onCompletion: @escaping (Bool) -> Void, onPCRChanged: @escaping (BootStateDiff) -> Void) {
        self.device = device
        self.context = context
        self.enroll = false
        self.bleManager = bleManager
        self.logger = logger
        self.onCompletion = onCompletion
        self.onPCRChanged = onPCRChanged
        
        ek = EkModel(EKCert: device.cert ?? Data(), EKPub: device.n!, EKExp: UInt(device.e), PCRExtend: device.secret != nil)
        bleManager.setOnMsgReadCallback(callback: self.onMsgRead)
        bleManager.setOnMsgWriteCallback(callback: self.onMsgWrite)
        startProtocol()
    }
    
    init(enrollData: EnrollData, context: NSManagedObjectContext, bleManager: BLEManager, logger: Logger, onCompletion: @escaping (Bool) -> Void, onPCRChanged: @escaping (BootStateDiff) -> Void) {
        self.enrollData = enrollData
        self.context = context
        self.enroll = true
        self.bleManager = bleManager
        self.logger = logger
        self.onCompletion = onCompletion
        self.onPCRChanged = onPCRChanged

        bleManager.setOnMsgReadCallback(callback: self.onMsgRead)
        bleManager.setOnMsgWriteCallback(callback: self.onMsgWrite)
        startProtocol()
    }
    
    private func startProtocol() {
        resumeProtocol()
    }
    
    func resumeProtocol(with msg: Data? = nil, at state: ProtoState? = nil) {
        if let s = state {
            currentState = s
        }
        switch currentState {
        case .uuid:
            do {
                let encoder = CodableCBOREncoder()
                let uid: UUID = enroll ? enrollData!.id : device!.uid!
                let encoded = try encoder.encode(ByteString(Bytes: Data.fromUUID(uid)))
                writeMsg(tag: "UUID", msg: encoded)
            }
            catch {
                print("An error occured while encoding CBOR: \(error)")
                return
            }
        case .start_encrypted_session:
            if let enrollData {
                cipher = EncryptedSession(key: enrollData.key)
            } else {
                do {
                    cipher = try EncryptedSession(for: device!.uid!)
                }
                catch {
                    print("An error occured while fetching encryption key from keychain: \(error)")
                    return
                }
            }
            resumeProtocol(at: .auth_read)
        case .auth_read:
            readMsg(tag: "auth nonce")
        case .auth_write:
            writeMsg(tag: "tweaked auth nonce", msg: msg!)
        case .after_auth:
            resumeProtocol(at: enroll ? .enroll : .ak_read)
        case .enroll:
            readMsg(tag: "EK pub and certificate")
        case .enroll_handle:
            do {
                let decoder = CodableCBORDecoder()
                ek = try decoder.decode(EkModel.self, from: msg!)
            }
            catch {
                print("An error occured while decoding CBOR: \(error)")
            }
            resumeProtocol(at: .ak_read)
        case .ak_read:
            readMsg(tag: "attestation key")
        case .make_credential:
            logger.push(log: Log("Generating activation credential"))
            encodedak = msg!
            let credentialBlob = Gomobile.GomobileMakeCredential(ek!.EKPub, Int(ek!.EKExp), encodedak, nil)
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
            logger.push(log: Log("Extract event log from attestation data"))
            guard let el = Gomobile.GomobileGetParsedEventLog(encodedpp, nil) else {
                logger.completeLast(success: false)
                return
            }
            eventlog =  String(decoding: el.raw!, as: UTF8.self)
            logger.completeLast(success: true)
            logger.push(log: Log("Replaying event log"))
            guard Gomobile.GomobileReplayEventLog(encodedpp, nil) else {
                logger.completeLast(success: false)
                return
            }
            logger.completeLast(success: true)
            resumeProtocol(at: .pcrs_read)
        case .pcrs_read:
            logger.push(log: Log("Extract PCRs from attestation data"))
            guard let wp = Gomobile.GomobileGetPCRs(encodedpp, nil) else {
                logger.completeLast(success: false)
                return
            }
            encodedPCRs = wp.data
            logger.completeLast(success: true)
            if enroll {
                if ek?.PCRExtend == true {
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
            logger.push(log: Log("Decoding PCRs"))
            let pcrs: [PCR]
            let refs: [PCR]
            do {
                let decoder = CodableCBORDecoder()
                pcrs = try decoder.decode([PCR].self, from: encodedPCRs!)
                refs = try decoder.decode([PCR].self, from: device!.encoded_pcrs!)
            }
            catch {
                logger.completeLast(success: false)
                return
            }
            logger.completeLast(success: true)
            logger.push(log: Log("Validating PCRs"))
            if pcrs.count != refs.count {
                logger.completeLast(success: false)
                return
            }
            logger.completeLast(success: true)
            var failure = false
            var diffs = [IdentifiablePCR]()
            // We want to get all pcrs result before raising the error, so we save the onError callback, and set it to void for now
            logger.setOnFailureCallback { }
            for i in 0..<refs.count {
                if Policy(device!.pcrpolicy).isPCRSet(index: refs[i].Index) {
                    if pcrs[i].Digest != refs[i].Digest {
                        logger.push(log: Log("PCR\(refs[i].Index) - \(PCRBank[refs[i].DigestAlg] ?? "Unknown")", success: false))
                        failure = true
                        diffs.append(IdentifiablePCR(id: UUID(), pcr: pcrs[i]))
                    }
                }
            }
            if failure {
                let diffStr = Gomobile.GomobileGetDiff(device!.eventlog!, eventlog!)
                let diff = BootStateDiff(PCRs: diffs, eventLogDiff: String(decoding: diffStr!.raw!, as: UTF8.self))
                onPCRChanged(diff)
                return
            } else {
                attestationResponse = AttestationResponse(err: false, Secret: device!.secret ?? Data())
            }
            resumeProtocol(at: .attestation_response)
        case .user_decision:
            logger.push(log: Log("Read user decision"))
            guard let m = msg else {
                logger.completeLast(success: false)
                return
            }
            logger.completeLast(success: true)
            if m[1] == 1 {
                logger.push(log: Log("Updating attester information"))
                guard let d = device else {
                    logger.completeLast(success: false)
                    return
                }
                d.encoded_pcrs = encodedPCRs
                d.eventlog = eventlog
                do {
                    try context.save()
                } catch {
                    logger.completeLast(success: false)
                    return
                }
                logger.completeLast(success: true)
            }
            if m[0] == 1 {
                attestationResponse = AttestationResponse(err: false, Secret: device!.secret ?? Data())
            } else {
                attestationResponse = AttestationResponse(err: true, Secret: Data())
            }
            resumeProtocol(at: .attestation_response)
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
            if enroll {
                saveNewDevice()
            }
            onCompletion(enroll || attestationResponse!.err == false)
        }
    }
    
    private func saveNewDevice() {
        logger.push(log: Log("Saving new attester entry"))
        do {
            try cipher?.storeKey(for: enrollData!.id)
        } catch {
            print(error.localizedDescription)
            logger.completeLast(success: false)
            return
        }
        device = Device(context: context)
        guard let d = device else {
            logger.completeLast(success: false)
            return
        }
        d.uid = enrollData!.id
        d.creation_time = Date.now
        d.last_attestation_time = d.creation_time
        d.name = Name.generate()
        d.pcrpolicy = Policy(.strict).value
        d.e = Int64(ek!.EKExp)
        d.encoded_pcrs = encodedPCRs
        d.n = ek!.EKPub
        d.cert = ek!.EKCert
        d.secret = secret
        d.eventlog = eventlog
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
        var data = msg
        if let c = cipher {
            data = try! c.encrypt(msg)
        }
        bleManager.setMessageTag(tag)
        bleManager.sendMsg(msg: data)
    }
    
    private func onMsgRead(msg: Data) {
        var data = msg
        if let c = cipher {
            data = try! c.decrypt(cipher: msg)
        }
        currentState = currentState.next()
        resumeProtocol(with: data)
    }
    
    private func onMsgWrite() {
        currentState = currentState.next()
        resumeProtocol()
    }
}
