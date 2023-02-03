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
                ContentCard(title: "Name", content: device.name ?? "", copiable: false)
                    .padding(.bottom, 10)
                ContentCard(title: "Creation", content: device.creation_time?.formatted() ?? "", copiable: false)
                    .padding(.bottom, 10)
                ContentCard(title: "Last attestation", content: device.last_attestation_time?.formatted() ?? "", copiable: false)
                    .padding(.bottom, 10)
                ContentCard(title: "UUID", content: (device.addr ?? "").trimmingCharacters(in: .newlines), copiable: true)
                    .padding(.bottom, 10)
                ContentCard(title: "PCR policy", content: Policy(device.pcrpolicy).toString(), copiable: false)
                    .padding(.bottom, 10)
                if device.cert != nil && device.cert!.count > 0 {
                    ContentCard(title: "EK Certificate", content: parseCertificate(from: device.cert!), copiable: true)
                        .padding(.bottom, 10)
                } else {
                    ContentCard(title: "EK Public", content: "\((device.n ?? Data()).base64EncodedString()) \(device.e)", copiable: true)
                        .padding(.bottom, 10)
                }
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
                    Text("Edit")
                })
                .sheet(isPresented: $isEditing, content: {
                    DeviceEditView(newName: device.name!, device: device, policy: Policy(device.pcrpolicy).toBoolArray())
                })
            }
        }
        .background(Color(UIColor.systemGroupedBackground))
    }
    
    private func parseCertificate(from data: Data) -> String {
        if data.isEmpty {
            return "N/A"
        }
        return "-----BEGIN CERTIFICATE-----\n" + data.base64EncodedString() + "\n-----END CERTIFICATE-----"
    }
    
}

struct DeviceView_Previews: PreviewProvider {
    static var previews: some View {
        DeviceView(device: Device())
    }
}
