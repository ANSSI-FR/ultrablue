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
}
