//
//  ContentView.swift
//  ultrablue
//
//  Created by loic buckwell on 09/07/2022.
//

import SwiftUI

struct ContentView: View {
    @Environment(\.managedObjectContext) private var viewContext

    @FetchRequest(
        sortDescriptors: [NSSortDescriptor(keyPath: \Item.timestamp, ascending: true)],
        animation: .default)
    private var items: FetchedResults<Item>

    var body: some View {
        TabView {
            DeviceListView()
            .tabItem {
                Text("Devices")
                Image(systemName: "rectangle.stack")
            }
            SettingsView()
            .tabItem {
                Text("Settings")
                Image(systemName: "gearshape")
            }
        }
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        Group {
            ContentView().environment(\.managedObjectContext, PersistenceController.preview.container.viewContext)
            ContentView().environment(\.managedObjectContext, PersistenceController.preview.container.viewContext)
            ContentView().environment(\.managedObjectContext, PersistenceController.preview.container.viewContext)
        }
    }
}
