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
    @StateObject var logger = Logger()
    var device: Device?
    @State var bleManager: BLEManager

    var body: some View {
        NavigationView {
            ScrollView {
                VStack {
                    ForEach(logger.lines) { log in
                        LogView(log: log)
                    }
                    Spacer()
                }
                .padding(.top, 10)
            }
            .navigationBarTitle("Attestation", displayMode: .inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(action: {
                        isActionSheetPresented = true
                    }, label: {
                        Text("Cancel")
                    })
                }
            }
        }
        .onAppear() {
            logger.setOnFailureCallback({
                bleManager.shutdown(err: true)
                self.isAlertPresented.toggle()
            })
            bleManager.setLogger(logger: logger)
            runAttestation()
        }
        .onDisappear() {
            logger.clear()
        }
        .alert(isPresented: $isAlertPresented) { () -> Alert in
            Alert(
                title: Text("Error"),
                message: Text("The attestation failed due to a communication error. This doesn't means the boot state changed."),
                primaryButton: .default(Text("Ok")),
                secondaryButton: .default(Text("Retry"), action: {
                    logger.clear()
                    logger.setOnFailureCallback {
                        bleManager.shutdown(err: true)
                        self.isAlertPresented.toggle()
                    }
                    runAttestation()
                })
            )
        }
        .confirmationDialog("Cancel attestation", isPresented: $isActionSheetPresented) {
            // TODO: replace attestation with enrollment when appropriate
            Button("Cancel attestation", role: .destructive) {
                bleManager.shutdown(err: true)
                logger.clear()
                dismiss()
            }
        }
        // TODO: Present the confirmation dialog when the feature is added by Apple.
        .interactiveDismissDisabled()
        .confettiCannon(counter: $confettiCounter, num: 150, rainHeight: 600, openingAngle: Angle(degrees: 0), closingAngle: Angle(degrees: 360), radius: 400)
    }
    
    func runAttestation() {
        bleManager.searchForAttestingDevice(onAttesterFound: {
            let _ = UltrablueProtocol(device: device, context: viewContext, bleManager: bleManager, logger: logger, onSuccess: {
                bleManager.shutdown()
                confettiCounter += 1
            })
        })
    }

}
