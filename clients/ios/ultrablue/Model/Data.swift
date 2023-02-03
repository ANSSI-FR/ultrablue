//
//  Data.swift
//  ultrablue
//
//  Created by loic buckwell on 19/01/2023.
//

import Foundation

extension Data {
    
    static func fromHexString(_ str: String) throws -> Data {
        // String validation
        if !str.count.isMultiple(of: 2) {
            throw DataError.InvalidArgumentError(desc: "Odd hex string length")
        }
        for c in str {
            if !c.isHexDigit {
                throw DataError.InvalidArgumentError(desc: "Invalid hex digit in string")
            }
        }
        
        // Get the UTF8 characters of this string
        let chars = Array(str.utf8)

        // Keep the bytes in an UInt8 array and later convert it to Data
        var bytes = [UInt8]()
        bytes.reserveCapacity(str.count / 2)

        // It is a lot faster to use a lookup map instead of strtoul
        let map: [UInt8] = [
          0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, // 01234567
          0x08, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 89:;<=>?
          0x00, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x00, // @ABCDEFG
          0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00  // HIJKLMNO
        ]

        // Grab two characters at a time, map them and turn it into a byte
        for i in stride(from: 0, to: str.count, by: 2) {
          let index1 = Int(chars[i] & 0x1F ^ 0x10)
          let index2 = Int(chars[i + 1] & 0x1F ^ 0x10)
          bytes.append(map[index1] << 4 | map[index2])
        }
        return Data(bytes)
    }
    
    static func fromUUID(_ uid: UUID) -> Data {
        var data = Data(count: 16)
        
        data[0] = uid.uuid.0
        data[1] = uid.uuid.1
        data[2] = uid.uuid.2
        data[3] = uid.uuid.3
        data[4] = uid.uuid.4
        data[5] = uid.uuid.5
        data[6] = uid.uuid.6
        data[7] = uid.uuid.7
        data[8] = uid.uuid.8
        data[9] = uid.uuid.9
        data[10] = uid.uuid.10
        data[11] = uid.uuid.11
        data[12] = uid.uuid.12
        data[13] = uid.uuid.13
        data[14] = uid.uuid.14
        data[15] = uid.uuid.15
        
        return data
    }
    
    private enum DataError: Error {
        case InvalidArgumentError(desc: String)
    }
    
}
