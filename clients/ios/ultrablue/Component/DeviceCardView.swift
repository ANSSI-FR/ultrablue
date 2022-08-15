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
    @State var bleManager: BLEManager
        
    @StateObject var device: Device
    
    var body: some View {
        HStack(alignment: .top) {
            VStack(alignment: .leading) {
                HStack(alignment: .top) {
                    if device.secret != nil && device.secret!.count > 0 {
                        Image(systemName: "key.fill")
                            .padding(.top, 5)
                            .foregroundColor(.gray)
                    }
                    Text(device.name ?? "")
                        .font(.system(size: 24))
                        .foregroundColor(Color(UIColor.label))
                        .bold()
                }
                .padding(.leading, 20)
                .padding(.top, 20)
                .padding(.bottom, 2)
                Text("Last attestation:\n" + (device.last_attestation_time?.formatted() ?? "--/--/--"))
                    .font(.system(size: 16, design: .monospaced))
                    .foregroundColor(Color(UIColor.label))
                    .padding(.bottom, 10)
                    .padding(.leading, 20)
            }
            Spacer()
            VStack(alignment: .trailing) {
                Button(action: {
                    runAttestation()
                }) {
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
        .sheet(isPresented: $showProtocolView, content: {
            ProtocolView(device: device, bleManager: bleManager)
        })
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
