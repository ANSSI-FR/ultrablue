//
//  ProtocolView.swift
//  ultrablue
//
//  Created by loic buckwell on 16/07/2022.
//

import SwiftUI

struct ProtocolView: View {
    var device: Device?
    
    var body: some View {
        Text("Protocol View")
            .navigationBarTitle(device != nil ? "Attestation" : "Enrollment", displayMode: .inline)
    }
}

struct ProtocolView_Previews: PreviewProvider {
    static var previews: some View {
        ProtocolView()
    }
}
