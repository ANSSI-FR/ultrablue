//
//  Policy.swift
//  ultrablue
//
//  Created by loic buckwell on 17/07/2022.
//

import Foundation

class Policy {
    var value: Int16
    
    enum toughness {
        case strict, permissive
    }
    
    init(_ toughness: toughness = .strict) {
        switch toughness {
        case .strict:
            value = Int16(bitPattern: UInt16.max)
        case .permissive:
            value = Int16(bitPattern: UInt16.min)
        }
    }
    
    init(_ val: Int16) {
        value = val
    }
    
    init(from values: [Bool]) {
        if values.count == 10 {
            let policy = Policy(.permissive)
            for i in 0..<10 {
                if values[i] == true {
                    policy.setPCR(index: i)
                }
            }
            value = policy.value
        } else {
            value = Int16(bitPattern: UInt16.max)
        }
    }
    
    func setPCR(index: Int) {
        if index >= 0 && index < 10 {
            value = value | (1 << index)
        }
    }
    
    func unsetPCR(index: Int) {
        if index >= 0 && index < 10 {
            value = value ^ (1 << index)
        }
    }
    
    func isPCRSet(index: Int) -> Bool {
        if index >= 0 && index < 10 {
            return (value & (1 << index)) != 0
        }
        return false
    }
    
    func print() {
        Swift.print(String(value, radix: 2))
    }
    
    func toBoolArray() -> [Bool] {
        var values = [Bool]()
        
        for i in 0..<10 {
            values.append(self.isPCRSet(index: i) ? true : false)
        }
        return values
    }
    
    func toString() -> String {
        var pcrs = [String]()
        
        for i in 0..<10 {
            if self.isPCRSet(index: i) {
                pcrs.append(String(i))
            } else {
                pcrs.append("-")
            }
        }
        return pcrs.joined(separator: "  ")
    }
}
