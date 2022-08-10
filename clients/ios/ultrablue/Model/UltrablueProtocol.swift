//
//  Protocol.swift
//  ultrablue
//
//  Created by loic buckwell on 22/07/2022.
//

import Foundation
import SwiftCBOR

extension CaseIterable where Self: Equatable {
    private var allCases: AllCases { Self.allCases }
    var next: Self {
        let index = allCases.index(after: allCases.firstIndex(of: self)!)
        guard index != allCases.endIndex else { return allCases.first! }
        return allCases[index]
    }
}

enum ProtoState: CaseIterable {
    case enroll, enroll_handle, auth_read, auth_write, ak_read, ak_handle, end
    
    func next() -> ProtoState {
        return self.next
    }
}

class UltrablueProtocol {
    private var logger: Logger
    private var currentState: ProtoState
    private var bleManager: BLEManager
    private var onSuccess: () -> Void
    
    private var enrollData: EnrollDataModel? = nil
    private var authNonce: Data? = nil
    
    init(bleManager: BLEManager, logger: Logger, onSuccess: @escaping () -> Void) {
        self.bleManager = bleManager
        self.logger = logger
        self.currentState = .enroll
        self.onSuccess = onSuccess
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
        case .ak_handle:
            
            resumeProtocol(at: .end)
        case .end:
            onSuccess()
        }
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
