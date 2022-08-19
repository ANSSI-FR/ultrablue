//
//  String.swift
//  ultrablue
//
//  Created by loic buckwell on 27/07/2022.
//

import Foundation

extension String {
    func replaceLast(with ch: Character) -> String {
        var newString = self
        if !newString.isEmpty {
            newString.removeLast()
            newString.append(ch)
        }
        return newString
    }
    
    func splitLines() -> [IdentifiableSubstring] {
        var identifiableLines = [IdentifiableSubstring]()
        let lines = self.split(separator: "\n")
        for line in lines {
            identifiableLines.append(IdentifiableSubstring(id: UUID(), raw: line))
        }
        return identifiableLines
    }
}

struct IdentifiableSubstring: Identifiable {
    let id: UUID
    let raw: Substring
}
