//
//  ProtocolView.swift
//  ultrablue
//
//  Created by loic buckwell on 16/07/2022.
//

import SwiftUI

struct ProtocolView: View {
    @State var logger = Logger()
    var device: Device?
    
    var body: some View {
        LoggerView(logger: $logger)
    }
}

struct ProtocolView_Previews: PreviewProvider {
    static var previews: some View {
        ProtocolView(device: nil)
    }
}
