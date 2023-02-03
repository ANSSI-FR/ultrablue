//
//  Conversion.swift
//  ultrablue
//
//  Created by loic buckwell on 10/08/2022.
//

import Foundation

extension Data {
    
    func toUInt() -> UInt {
        let number = self.withUnsafeBytes { pointer in
            return pointer.load(as: Int32.self)
        }
        return UInt(number)
    }

}

extension UInt {
    
    func toLittleEndianData() -> Data {
        let val = UInt32(self)
        let array = withUnsafeBytes(of: val.littleEndian, Array.init)
        return Data(array)
    }
    
}
