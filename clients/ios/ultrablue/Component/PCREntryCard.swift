//
//  PCREntryCard.swift
//  ultrablue
//
//  Created by loic buckwell on 17/07/2022.
//

import SwiftUI

struct PCREntryCard: View {
    var index: Int
    var description: String
    @Binding var isOn: Bool
    
    var body: some View {
        HStack(alignment: .top) {
            VStack(alignment: .leading) {
                Text("PCR \(index)")
                    .font(.system(size: 19))
                    .foregroundColor(Color(UIColor.label))
                    .bold()
                    .padding(.top, 5)
                Text(description)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .font(.system(size: 16))
                    .foregroundColor(Color(UIColor.systemGray))
            }
            VStack(alignment: .trailing) {
                Toggle(isOn: $isOn) {
                    
                }
                .padding(.top, 10)
                .tint(Color.accentColor)
                .frame(maxWidth: 60)
            }
        }
        
    }
}

struct PCREntryCard_Previews: PreviewProvider {
    @State static var isOn = false
    static var previews: some View {
        PCREntryCard(index: 3, description: "fasdfhkl ashdfkashd fkj ahqiofe hquw ihvu", isOn: $isOn)
    }
}
