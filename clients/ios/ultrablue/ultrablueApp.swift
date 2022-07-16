//
//  ultrablueApp.swift
//  ultrablue
//
//  Created by loic buckwell on 09/07/2022.
//

import SwiftUI

@main
struct ultrablueApp: App {
    let persistenceController = PersistenceController.shared

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(\.managedObjectContext, persistenceController.container.viewContext)
        }
    }
}
