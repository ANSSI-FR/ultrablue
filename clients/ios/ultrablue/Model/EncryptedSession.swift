//
//  EncryptedSession.swift
//  ultrablue
//
//  Created by loic buckwell on 19/01/2023.
//

import Foundation
import CryptoKit

class EncryptedSession {
    let key: SymmetricKey
    
    init(key: SymmetricKey) {
        self.key = key
    }
    
    init(for id: UUID) throws {
        let tag = "com.ANSSI.ultrablue.key-\(id.uuidString)".data(using: .utf8)!
        let getquery: [String: Any] = [kSecClass as String: kSecClassKey,
                                       kSecAttrApplicationTag as String: tag,
                                       kSecAttrKeyType as String: kSecAttrKeyClassSymmetric,
                                       kSecReturnData as String: true]
        var result: AnyObject?
        let status = SecItemCopyMatching(getquery as CFDictionary, &result)
        guard status == errSecSuccess else { throw EncryptedSessionError.keychainError }
        self.key = SymmetricKey(data: result as! Data)
    }
    
    func encrypt(_ msg: Data) throws -> Data {
        let iv = AES.GCM.Nonce()
        let sb = try AES.GCM.seal(msg, using: key, nonce: iv)
        if let cipher = sb.combined {
            return cipher
        }
        throw EncryptedSessionError.combineError
    }
    
    func decrypt(cipher: Data) throws -> Data {
        let sb = try AES.GCM.SealedBox(combined: cipher)
        let plaintext = try AES.GCM.open(sb, using: self.key)
        return plaintext
    }
    
    func storeKey(for id: UUID) throws {
        let keydata = self.key.withUnsafeBytes {
            return Data(Array($0))
        }
        let tag = "com.ANSSI.ultrablue.key-\(id.uuidString)".data(using: .utf8)!
        let addquery: [String: Any] = [kSecClass as String: kSecClassKey,
                                       kSecAttrApplicationTag as String: tag,
                                       kSecAttrKeyType as String: kSecAttrKeyClassSymmetric,
                                       kSecValueData as String: keydata]
        let status = SecItemAdd(addquery as CFDictionary, nil)
        guard status == errSecSuccess else { throw EncryptedSessionError.keychainError }
    }
    
}

extension EncryptedSession {
    enum EncryptedSessionError: Error {
        case combineError
        case keychainError
    }
}
