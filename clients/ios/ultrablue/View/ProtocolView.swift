//
//  ProtocolView.swift
//  ultrablue
//
//  Created by loic buckwell on 16/07/2022.
//

import SwiftUI
import Combine
import CoreBluetooth
import ConfettiSwiftUI

struct ProtocolView: View {
    @Environment(\.dismiss) var dismiss: DismissAction
    @Environment(\.managedObjectContext) private var viewContext
    @State private var isAlertPresented = false
    @State private var isActionSheetPresented = false
    @State private var confettiCounter = 0
    @State private var protocolSuccess: Bool? = nil
    @StateObject var logger = Logger()
    var device: Device?
    @State var bleManager: BLEManager
    
    @Namespace var bottomID
        
    var body: some View {
        NavigationView {
            ScrollViewReader { proxy in
                ScrollView {
                    VStack {
                        ForEach(logger.lines) { log in
                            LogView(log: log)
                        }
                        .onChange(of: self.logger.lines.count, perform: { _ in
                            proxy.scrollTo(bottomID)
                        })
                        Spacer()
                        Button("bottomID") { }
                        .id(bottomID)
                        .disabled(true)
                        .opacity(0)
                    }
                    .padding(.top, 10)
                }
            }
            .navigationBarTitle("Attestation", displayMode: .inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(action: {
                        if protocolSuccess != nil {
                            logger.clear()
                            dismiss()
                        } else {
                            isActionSheetPresented = true
                        }
                    }, label: {
                        Text(protocolSuccess != nil ? "Done" : "Cancel")
                    })
                }
                ToolbarItem(placement: .automatic) {
                    Button(action: {
                        reset()
                        runAttestation()
                    }, label: {
                        Text("Retry")
                    })
                    .disabled(protocolSuccess != false)
                }
            }
        }
        .onAppear() {
            bleManager.setLogger(logger: logger)
            reset()
            runAttestation()
        }
        .alert(isPresented: $isAlertPresented) { () -> Alert in
            Alert(
                title: Text("Error"),
                message: Text("The attestation failed due to a communication error. This doesn't means the boot state changed."),
                primaryButton: .default(Text("Ok")),
                secondaryButton: .default(Text("Retry"), action: {
                    reset()
                    runAttestation()
                })
            )
        }
        .confirmationDialog("Cancel attestation", isPresented: $isActionSheetPresented) {
            // TODO: replace attestation with enrollment when appropriate
            Button("Cancel attestation", role: .destructive) {
                logger.setOnFailureCallback { }
                bleManager.shutdown(err: true)
                logger.clear()
                dismiss()
            }
        }
        // TODO: Present the confirmation dialog when the feature is added by Apple.
        .interactiveDismissDisabled(protocolSuccess == nil)
        .confettiCannon(counter: $confettiCounter, num: 150, rainHeight: 600, openingAngle: Angle(degrees: 0), closingAngle: Angle(degrees: 360), radius: 400)
    }
    
    func runAttestation() {
        protocolSuccess = nil
        bleManager.searchForAttestingDevice(onAttesterFound: {
            let _ = UltrablueProtocol(device: device, context: viewContext, bleManager: bleManager, logger: logger, onCompletion: { success in
                bleManager.shutdown()
                protocolSuccess = success
                if success {
                    if device == nil {
                        dismiss()
                    } else {
                        device!.last_attestation_time = Date.now
                        do {
                            try viewContext.save()
                        } catch {
                            print("Error while updating device: \(error)")
                        }
                        confettiCounter += 1
                    }
                } else {
                    print("Printing event log diffs")
                }
            })
        })
    }
    
    func reset() {
        logger.clear()
        protocolSuccess = nil
        logger.setOnFailureCallback {
            bleManager.shutdown(err: true)
            isAlertPresented.toggle()
            protocolSuccess = false
        }
    }

}
