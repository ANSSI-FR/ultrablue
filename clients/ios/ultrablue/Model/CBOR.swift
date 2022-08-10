//
//  CBOR.swift
//  ultrablue
//
//  Created by loic buckwell on 10/08/2022.
//

import Foundation
import SwiftCBOR

struct EnrollDataModel: Codable {
    let N: Data
    let E: UInt
    let PCRExtend: Bool
}
