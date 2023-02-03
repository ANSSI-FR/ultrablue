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
import CryptoKit

struct ProtocolView: View {
    @Environment(\.dismiss) var dismiss: DismissAction
    @Environment(\.managedObjectContext) private var viewContext
    @State private var isAlertPresented = false
    @State private var isActionSheetPresented = false
    @State private var confettiCounter = 0
    @State private var ubProtocol: UltrablueProtocol? = nil
    @State private var protocolSuccess: Bool? = nil
    @State private var PCRPolicyError: Bool = false
    @State private var bootStateDiff: BootStateDiff? = nil
    @StateObject var logger = Logger()
    var device: Device?
    var enrollData: EnrollData?
    @State var bleManager: BLEManager
    
    @Namespace var bottomID
    
    init(device: Device, bleManager: BLEManager) {
        self.device = device
        self.enrollData = nil
        _bleManager = State(initialValue: bleManager)
    }
    
    init(enrollData: EnrollData, bleManager: BLEManager) {
        self.device = nil
        self.enrollData = enrollData
        _bleManager = State(initialValue: bleManager)
    }
        
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
                .sheet(isPresented: $PCRPolicyError, content: {
                    if let diff = bootStateDiff {
                        ScrollView {
                            VStack {
                                Text("Boot state has changed")
                                    .bold()
                                    .font(.system(size: 24))
                                    .foregroundColor(.red)
                                    .padding(20)
                                ContentHolderCard(title: "Modified PCR(s)", details: "Those values has been cryptographically verified and can be trusted.") {
                                    VStack(alignment: .leading) {
                                        ForEach(diff.PCRs) { pcr in
                                            Text("PCR\(pcr.pcr.Index) - \(PCRBank[pcr.pcr.DigestAlg] ?? "Unknown") bank:")
                                                .font(.system(size: 16))
                                                .bold()
                                            Text("\(PCRs[pcr.pcr.Index].desc)")
                                                .font(.system(size: 12))
                                                .padding(.leading, 10)
                                        }
                                    }
                                }
                                .padding(.horizontal, 10)
                                ContentHolderCard(title: "Event log diff", details: "Event logs can't be trusted as PCRs does, they are used for debug purpose only.") {
                                    ZStack(alignment: .topTrailing) {
                                        VStack(alignment: .leading) {
                                            ForEach(diff.eventLogDiff.splitLines()) { line in
                                                if (line.raw.starts(with: "@")) {
                                                    Text(line.raw)
                                                        .font(.system(size: 14))
                                                        .bold()
                                                } else {
                                                    Text(line.raw)
                                                        .font(.system(size: 14))
                                                        .foregroundColor(line.raw.starts(with: ">") ? .green : line.raw.starts(with: "<") ? .red : Color(UIColor.label))
                                                }
                                            }
                                        }
                                        .frame(maxWidth: .infinity)
                                        Button(action: {
                                            UIPasteboard.general.string = diff.eventLogDiff
                                        }) {
                                            Image(systemName: "square.on.square")
                                        }
                                        .padding(5)
                                    }
                                }
                                .padding(.horizontal, 10)
                                .padding(.bottom, 5)
                                Spacer()
                                Button("Trust and save new boot state") {
                                    PCRPolicyError = false
                                    resumeAttestation(trust: true, save: true)
                                }
                                    .frame(maxWidth: .infinity, minHeight: 45)
                                    .foregroundColor(.white)
                                    .background(.blue)
                                    .cornerRadius(7)
                                    .padding(.horizontal, 15)
                                Button("Trust boot state this time only") {
                                    PCRPolicyError = false
                                    resumeAttestation(trust: true, save: false)
                                }
                                    .frame(maxWidth: .infinity, minHeight: 45)
                                    .foregroundColor(.white)
                                    .background(.blue)
                                    .cornerRadius(7)
                                    .padding(.horizontal, 15)
                                Button("Mark as failed") {
                                    PCRPolicyError = false
                                    resumeAttestation(trust: false, save: false)
                                }
                                    .frame(maxWidth: .infinity, minHeight: 45)
                                    .foregroundColor(.white)
                                    .background(.red)
                                    .cornerRadius(7)
                                    .padding(.horizontal, 15)
                            }
                        }
                        .background(Color(UIColor.systemGroupedBackground))
                        .interactiveDismissDisabled(true)
                    }
                })
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
            let onCompletion: (Bool) -> () = { success in
                bleManager.shutdown()
                protocolSuccess = success
                if success {
                    if device == nil {
                        dismiss()
                    } else {
                        device!.last_attestation_time = Date.now
                        confettiCounter += 1
                    }
                }
                do {
                    try viewContext.save()
                } catch {
                    print("Error while updating device: \(error)")
                }
            }
            let onPCRChanged: (BootStateDiff) -> () = { diff in
                PCRPolicyError = true
                bootStateDiff = diff
            }
            
            if device != nil {
                ubProtocol = UltrablueProtocol(device: device!, context: viewContext, bleManager: bleManager, logger: logger, onCompletion: onCompletion, onPCRChanged: onPCRChanged)
            } else {
                ubProtocol = UltrablueProtocol(enrollData: enrollData!, context: viewContext, bleManager: bleManager, logger: logger, onCompletion: onCompletion, onPCRChanged: onPCRChanged)
            }
                
        })
    }
    
    func resumeAttestation(trust: Bool, save: Bool) {
        if let p = ubProtocol {
            p.resumeProtocol(with: Data([trust ? 1 : 0, save ? 1 : 0]), at: .user_decision)
        }
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
