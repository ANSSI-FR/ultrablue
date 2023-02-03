//
//  CBOR.swift
//  ultrablue
//
//  Created by loic buckwell on 10/08/2022.
//

import Foundation
import SwiftCBOR

struct EkModel: Codable {
    let EKCert: Data
    let EKPub: Data
    let EKExp: UInt
    let PCRExtend: Bool
}

struct EncryptedCredentialModel: Codable {
    let Credential: Data
    let Secret: Data
}

struct ByteString: Codable {
    let Bytes: Data
}

struct AttestationResponse: Codable {
    let err: Bool
    let Secret: Data
}

struct PCR: Codable {
    let Index: Int
    let Digest: Data
    let DigestAlg: UInt
}

struct IdentifiablePCR: Identifiable {
    var id: UUID
    var pcr: PCR
}
