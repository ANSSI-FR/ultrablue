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
    @State private var isAlertPresented = false
    @State private var confettiCounter = 0
    @StateObject var logger = Logger()
    var address: String
    @Binding var bleManager: BLEManager

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
                        print("Tried to cancel attestation")
                    }, label: {
                        Text("Cancel")
                    })
                }
            }
        }
        .onAppear() {
            logger.setOnFailureCallback({
                bleManager.shutdown()
                self.isAlertPresented.toggle()
            })
            bleManager.setLogger(logger: logger)
            runAttestation()
        }
        .onDisappear() {
            bleManager.shutdown()
            logger.clear()
        }
        .alert(isPresented: $isAlertPresented) { () -> Alert in
            Alert(
                title: Text("Error"),
                message: Text("The attestation failed due to a communication error. This doesn't means the boot state changed."),
                primaryButton: .default(Text("Ok")),
                secondaryButton: .default(Text("Retry"), action: {
                    logger.clear()
                    runAttestation()
                })
            )
        }
        .confettiCannon(counter: $confettiCounter, num: 150, rainHeight: 600, openingAngle: Angle(degrees: 0), closingAngle: Angle(degrees: 360), radius: 400)
    }
    
    func runAttestation() {
        bleManager.searchForAttestingDevice(onAttesterFound: {
            let _ = UltrablueProtocol(bleManager: bleManager, logger: logger, onSuccess: {
                bleManager.shutdown()
                confettiCounter += 1
            })
        })
    }
    
}
