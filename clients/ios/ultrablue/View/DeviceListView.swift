//
//  DeviceListView.swift
//  ultrablue
//
//  Created by loic buckwell on 15/07/2022.
//

import SwiftUI
import CodeScanner
import CoreData
import CryptoKit

struct DeviceListView: View {
    @Environment(\.managedObjectContext) private var viewContext
    @FetchRequest(sortDescriptors: [SortDescriptor(\.last_attestation_time, order: .reverse)])
    var devices: FetchedResults<Device>
    @State private var enrollData: EnrollData? = nil
    @State private var isShowingScanner = false
    @State private var bleManager = BLEManager()
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(alignment: .leading) {
                    ForEach(devices) { device in
                        DeviceCardView(bleManager: bleManager, device: device)
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
                    .sheet(item: $enrollData) { qr in
                        ProtocolView(enrollData: qr, bleManager: bleManager)
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
            do {
                enrollData = try EnrollData(Data(result.string.utf8))
            } catch {
                // TODO: Show an alert
                print("Invalid registration QR code: \(error)")
            }
        case .failure(let err):
            // TODO: Show an alert
           print("Scanning failed \(err.localizedDescription)")
        }
    }
    
}

struct DeviceListView_Previews: PreviewProvider {
    static var previews: some View {
        DeviceListView().environment(\.managedObjectContext, PersistenceController.preview.container.viewContext)
    }
}
