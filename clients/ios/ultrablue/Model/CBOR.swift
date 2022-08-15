//
//  CBOR.swift
//  ultrablue
//
//  Created by loic buckwell on 10/08/2022.
//

import Foundation
import SwiftCBOR

struct EnrollDataModel: Codable {
    let Cert: Data
    let N: Data
    let E: UInt
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
