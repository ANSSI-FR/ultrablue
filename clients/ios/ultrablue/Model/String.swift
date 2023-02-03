//
//  String.swift
//  ultrablue
//
//  Created by loic buckwell on 19/07/2022.
//

import Foundation

extension String {
    mutating func replaceLast(with char: Character) -> String {
        if !self.isEmpty {
            self.removeLast()
            self.insert(char, at: self.endIndex)
        }
        return self
    }
}
