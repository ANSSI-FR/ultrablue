//
//  DeviceCardView.swift
//  ultrablue
//
//  Created by loic buckwell on 14/07/2022.
//

import SwiftUI

struct FlatLinkStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
    }
}

struct DeviceCardView: View {
    @Environment(\.managedObjectContext) private var viewContext
    @State private var showProtocolView = false
    @State private var showDeviceView = false
        
    @StateObject var device: Device
    
    var body: some View {
        HStack(alignment: .top) {
            VStack(alignment: .leading) {
                Text(device.name ?? "")
                    .font(.system(size: 24))
                    .foregroundColor(Color(UIColor.label))
                    .bold()
                    .padding(.leading, 20)
                    .padding(.top, 20)
                    .padding(.bottom, 2)
                Text(device.addr ?? "")
                    .font(.system(size: 16, design: .monospaced))
                    .foregroundColor(Color(UIColor.label))
                    .padding(.bottom, 10)
                    .padding(.leading, 20)
            }
            Spacer()
            VStack(alignment: .trailing) {
                Button(action: runAttestation) {
                    Image(systemName: "play.fill")
                }
                .font(.system(size: 25))
                .padding(15)
            }
        }
        .background(Color(UIColor.secondarySystemGroupedBackground))
        .cornerRadius(10)
        .contextMenu {
            Button(action: runAttestation) {
                Text("Run attestation")
                Image(systemName: "play.fill")
            }
            Button(action: showDeviceInfo) {
                Text("Show device info")
                Image(systemName: "pencil")
            }
            Button(role: .destructive, action: deleteDevice) {
                Text("Delete device")
                Image(systemName: "trash.fill")
            }
        }
        .onTapGesture {
            showDeviceInfo()
        }
        NavigationLink(destination: ProtocolView(device: device), isActive: $showProtocolView) {
            EmptyView()
        }
        NavigationLink(destination: DeviceView(device: device), isActive: $showDeviceView) {
            EmptyView()
        }
    }
    
    private func runAttestation() {
        self.showProtocolView = true
    }
    
    private func showDeviceInfo() {
        self.showDeviceView = true
    }
    
    private func deleteDevice() {
        viewContext.delete(device)
        do {
            try viewContext.save()
        } catch {
            // TODO: Show an alert
            print("Failed to delete device")
        }
    }
}

struct DeviceCardView_Previews: PreviewProvider {
    static var previews: some View {
        DeviceCardView(device: Device())
    }
}
