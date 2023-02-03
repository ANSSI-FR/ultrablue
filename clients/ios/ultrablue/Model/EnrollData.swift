//
//  EnrollData.swift
//  ultrablue
//
//  Created by loic buckwell on 19/01/2023.
//

import Foundation
import CryptoKit

struct EnrollData: Identifiable {
    let id: UUID
    let mac: String // Unused, here for Android compatibility
    let key: SymmetricKey
    
    private struct RawQRData: Codable {
        let addr: String
        let key: String
    }
    
    init(_ data: Data) throws {
        let raw: RawQRData = try JSONDecoder().decode(RawQRData.self, from: data)
        let keyData = try Data.fromHexString(raw.key)
        
        self.mac = raw.addr
        self.key = SymmetricKey(data: keyData)
        self.id = UUID()
    }
}
