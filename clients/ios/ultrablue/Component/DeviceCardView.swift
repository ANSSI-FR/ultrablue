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
    @State private var showProtocolView = false
    @State private var showDeviceView = false
    
    var body: some View {
        HStack(alignment: .top) {
            VStack(alignment: .leading) {
                Text("Device Name")
                    .font(.system(size: 24))
                    .foregroundColor(Color(UIColor.label))
                    .bold()
                    .padding(.leading, 20)
                    .padding(.top, 20)
                Text("6f:3b:ff:0a:82:6c")
                    .font(.system(size: 16))
                    .foregroundColor(Color(UIColor.label))
                    .padding(.bottom, 20)
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
        .hoverEffect(.highlight)
        .onTapGesture {
            withAnimation(.linear) {
                showDeviceInfo()
            }
        }
        NavigationLink(destination: ProtocolView(), isActive: $showProtocolView) {
            EmptyView()
        }
        NavigationLink(destination: DeviceView(), isActive: $showDeviceView) {
            EmptyView()
        }
    }
    
    private func runAttestation() {
        print("run attestation")
        self.showProtocolView = true
    }
    
    private func showDeviceInfo() {
        print("show device info")
        self.showDeviceView = true
    }
    
    private func deleteDevice() {
        print("delete device")
    }
}

struct DeviceCardView_Previews: PreviewProvider {
    static var previews: some View {
        DeviceCardView()
    }
}
