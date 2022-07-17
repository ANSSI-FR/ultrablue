//
//  DeviceEditView.swift
//  ultrablue
//
//  Created by loic buckwell on 17/07/2022.
//

import SwiftUI
import Introspect

struct DeviceEditView: View {
    @Environment(\.managedObjectContext) private var viewContext
    @Environment(\.dismiss) var dismiss: DismissAction
    @State var newName: String
    @StateObject var device: Device
    
    var body: some View {
        NavigationView {
            HStack(alignment: .top) {
                VStack(alignment: .center) {
                    VStack(alignment: .center) {
                        Image(systemName: "qrcode")
                            .font(.system(size: 60))
                            .padding(15)
                        TextField("Device name", text: $newName)
                            .font(.system(size: 23, design: .rounded).weight(.bold))
                            .frame(height: 50)
                            .background(Color(UIColor.systemGroupedBackground))
                            .cornerRadius(10)
                            .padding([.horizontal, .bottom], 15)
                            .multilineTextAlignment(.center)
                            .introspectTextField { tf in
                                tf.becomeFirstResponder()
                                tf.clearButtonMode = .whileEditing
                            }
                    }
                    .background(Color(UIColor.secondarySystemGroupedBackground))
                    .cornerRadius(10)
                    .padding(10)
                    Button(action: {
                        viewContext.delete(device)
                        dismiss()
                    }) {
                       Text("Delete")
                            .foregroundColor(.white)
                            .font(.system(size: 23))
                            .frame(maxWidth: .infinity, minHeight: 60)
                            .background(Color(UIColor.systemRed))
                            .cornerRadius(10)
                    }
                    .padding(10)
                    Spacer()
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .padding(10)
            .navigationBarTitle("Edit", displayMode: .inline)
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    Button(action: {
                        device.name = newName
                        print(newName)
                        do {
                            try viewContext.save()
                        } catch {
                            print("An error occured while updating device name")
                        }
                        dismiss()
                    }, label: {
                        Text("Done")
                            .fontWeight(.bold)
                    })
                    .disabled(newName.count == 0)
                }
                ToolbarItem(placement: .cancellationAction) {
                    Button(action: {
                        dismiss()
                    }, label: {
                        Text("Cancel")
                    })
                }
            }
            .background(Color(UIColor.systemGroupedBackground))
        }
    }
}
