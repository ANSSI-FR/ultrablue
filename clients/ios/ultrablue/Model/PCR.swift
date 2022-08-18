//
//  PCR.swift
//  ultrablue
//
//  Created by loic buckwell on 17/07/2022.
//

import Foundation

let PCRBank: [UInt: String] = [
    0x3: "Sha1",
    0x5: "Sha256",
]

let PCRs = [
    PCRInfo(0, "Core System Firmware executable code"),
    PCRInfo(1, "Core System Firmware data"),
    PCRInfo(2, "Extended or pluggable executable code"),
    PCRInfo(3, "Extended or pluggable firmware data"),
    PCRInfo(4, "Boot Manager Code and Boot Attempts"),
    PCRInfo(5, "Boot Manager Configuration and Data"),
    PCRInfo(6, "Resume from S4 and S5 Power State Events"),
    PCRInfo(7, "Secure Boot State"),
    PCRInfo(8, "Hash of the kernel command line"),
    PCRInfo(9, "Hash of the initrd"),
    PCRInfo(10, "Application reserved PCR"),
    PCRInfo(11, "Application reserved PCR"),
    PCRInfo(12, "Application reserved PCR"),
    PCRInfo(13, "Application reserved PCR"),
    PCRInfo(14, "Application reserved PCR"),
    PCRInfo(15, "Application reserved PCR"),
    PCRInfo(16, "Application reserved PCR"),
    PCRInfo(17, "Application reserved PCR"),
    PCRInfo(18, "Application reserved PCR"),
    PCRInfo(19, "Application reserved PCR"),
    PCRInfo(20, "Application reserved PCR"),
    PCRInfo(21, "Application reserved PCR"),
    PCRInfo(22, "Application reserved PCR"),
    PCRInfo(23, "Application reserved PCR"),
]

class PCRInfo : Identifiable {
    var index: Int
    var desc: String
    
    init(_ index: Int, _ desc: String) {
        self.index = index
        self.desc = desc
    }
}
