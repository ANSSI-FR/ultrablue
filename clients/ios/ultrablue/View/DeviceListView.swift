//
//  DeviceListView.swift
//  ultrablue
//
//  Created by loic buckwell on 15/07/2022.
//

import SwiftUI
import CodeScanner

struct DeviceListView: View {
    @State private var isShowingScanner = false
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(alignment: .leading) {
                    DeviceCardView()
                        .padding(.bottom, 3)
                    DeviceCardView()
                        .padding(.bottom, 3)
                    DeviceCardView()
                        .padding(.bottom, 3)
                    DeviceCardView()
                        .padding(.bottom, 3)
                    DeviceCardView()
                        .padding(.bottom, 3)
                    DeviceCardView()
                        .padding(.bottom, 3)
                    DeviceCardView()
                        .padding(.bottom, 3)
                    DeviceCardView()
                        .padding(.bottom, 3)
                }
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
           print("Success with \(result.string)")
        case .failure(let err):
           print("Scanning failed \(err.localizedDescription)")
        }
    }
}

struct DeviceListView_Previews: PreviewProvider {
    static var previews: some View {
        DeviceListView()
    }
}
