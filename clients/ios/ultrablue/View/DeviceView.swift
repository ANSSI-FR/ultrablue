//
//  DeviceView.swift
//  ultrablue
//
//  Created by loic buckwell on 15/07/2022.
//

import SwiftUI

struct DeviceView: View {
    var device: Device
    
    var body: some View {
        Text("Device View")
            .navigationBarTitle(device.name!, displayMode: .inline)
    }
}

struct DeviceView_Previews: PreviewProvider {
    static var previews: some View {
        DeviceView(device: Device())
    }
}
