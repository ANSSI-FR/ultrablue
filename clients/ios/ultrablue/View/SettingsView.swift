//
//  SettingsView.swift
//  ultrablue
//
//  Created by loic buckwell on 15/07/2022.
//

import SwiftUI

struct SettingsView: View {
    var body: some View {
        NavigationView {
            Text("App settings tab")
                .navigationTitle("Settings")
        }
    }
}

struct SettingsView_Previews: PreviewProvider {
    static var previews: some View {
        SettingsView()
    }
}
