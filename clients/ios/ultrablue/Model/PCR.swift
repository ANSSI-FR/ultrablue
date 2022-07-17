//
//  PCR.swift
//  ultrablue
//
//  Created by loic buckwell on 17/07/2022.
//

import Foundation

let PCRs = [
    PCR(0, "Core System Firmware executable code"),
    PCR(1, "Core System Firmware data"),
    PCR(2, "Extended or pluggable executable code"),
    PCR(3, "Extended or pluggable firmware data"),
    PCR(4, "Boot Manager Code and Boot Attempts"),
    PCR(5, "Boot Manager Configuration and Data"),
    PCR(6, "Resume from S4 and S5 Power State Events"),
    PCR(7, "Secure Boot State"),
    PCR(8, "Hash of the kernel command line"),
    PCR(9, "Hash of the initrd"),
]

class PCR : Identifiable {
    var index: Int
    var desc: String
    
    init(_ index: Int, _ desc: String) {
        self.index = index
        self.desc = desc
    }
}
