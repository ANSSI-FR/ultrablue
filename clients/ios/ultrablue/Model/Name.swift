//
//  Name.swift
//  ultrablue
//
//  Created by loic buckwell on 16/07/2022.
//

import Foundation

class Name {
    private static let animals = [
        "Cat", "Dog", "Horse", "Ant", "Butterfly", "Duck", "Monkey", "Parrot"
    ]
    private static let characteristics = [
        "Black", "White", "Red", "Blue", "Green", "Fat", "Small", "Punky", "Rocky", "Jelly", "Great", "Smart", "Funny", "Stupid"
    ]
    
    static func generate() -> String {
        let animal = animals.randomElement()!
        let characteristic = characteristics.randomElement()!

        return "\(characteristic) \(animal)"
    }
}
