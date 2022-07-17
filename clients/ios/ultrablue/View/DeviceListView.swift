//
//  DeviceListView.swift
//  ultrablue
//
//  Created by loic buckwell on 15/07/2022.
//

import SwiftUI
import CodeScanner
import CoreData

struct DeviceListView: View {
    @Environment(\.managedObjectContext) private var viewContext
    @FetchRequest(entity: Device.entity(), sortDescriptors: [])
    var devices: FetchedResults<Device>
    @State private var isShowingScanner = false
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(alignment: .leading) {
                    ForEach(devices) { device in
                        DeviceCardView(device: device)
                            .padding(.bottom, 3)
                    }
                }
                .frame(
                    minWidth: 0,
                    maxWidth: .infinity,
                    minHeight: 0,
                    maxHeight: .infinity,
                    alignment: .topLeading
                )
                .padding(15)
            }
            .background(Color(UIColor.systemGroupedBackground))
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    Button(action: {
                        self.isShowingScanner = true
                    }, label: {
                        Image(systemName: "plus")
                    })
                    .sheet(isPresented: $isShowingScanner) {
                        CodeScannerView(codeTypes: [.qr], showViewfinder: true, completion: self.handleScan)
                    }
                }
            }
            .navigationTitle("Devices")
        }
    }
    
    private func handleScan(result: Result<CodeScanner.ScanResult, CodeScanner.ScanError>) {
        self.isShowingScanner = false
        switch result {
        case .success(let result):
            if isValidRegistrationData(data: result.string) {
                let device = Device(context: viewContext)
                device.id = UUID()
                device.addr = result.string
                device.name = Name.generate()
                device.pcrpolicy = Policy(.strict).value
                do {
                    try viewContext.save()
                } catch {
                    // TODO: Show an alert
                    print("An error occured")
                }
            } else {
                // TODO: Show an alert
                print("Invalid registration QR code")
            }
        case .failure(let err):
            // TODO: Show an alert
           print("Scanning failed \(err.localizedDescription)")
        }
    }
    
    private func isValidRegistrationData(data: String) -> Bool {
        // TODO: Parse MAC address
        return data.count == 18
    }
}

struct DeviceListView_Previews: PreviewProvider {
    static var previews: some View {
        DeviceListView().environment(\.managedObjectContext, PersistenceController.preview.container.viewContext)
    }
}
