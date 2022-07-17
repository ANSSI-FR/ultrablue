//
//  DeviceView.swift
//  ultrablue
//
//  Created by loic buckwell on 15/07/2022.
//

import SwiftUI

struct DeviceView: View {
    @State private var isEditing = false
    @StateObject var device: Device
    
    var body: some View {
        ScrollView {
            VStack {
                ContentCard(title: "name", content: device.name!)
                    .padding(.bottom, 10)
                ContentCard(title: "address", content: device.addr!.trimmingCharacters(in: .newlines))
                    .padding(.bottom, 10)
                ContentCard(title: "Endorsement key", content: "N/A")
                    .padding(.bottom, 10)
                ContentCard(title: "Certificate", content: "N/A")
                    .padding(.bottom, 10)
                Spacer()
            }
        }
        .padding(10)
        .navigationBarTitle("Device info", displayMode: .inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button(action: {
                    isEditing = true
                }, label: {
                    Image(systemName: "pencil")
                })
                .sheet(isPresented: $isEditing, content: {
                    DeviceEditView(newName: device.name!, device: device)
                })
            }
        }
        .background(Color(UIColor.systemGroupedBackground))
    }
}

struct DeviceView_Previews: PreviewProvider {
    static var previews: some View {
        DeviceView(device: Device())
    }
}
