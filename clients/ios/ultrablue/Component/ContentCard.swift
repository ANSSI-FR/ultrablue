//
//  ContentCard.swift
//  ultrablue
//
//  Created by loic buckwell on 16/07/2022.
//

import SwiftUI
import UniformTypeIdentifiers

struct ContentCard: View {
    var title: String
    var content: String
    var copiable: Bool
    
    var body: some View {
        HStack(alignment: .top) {
            VStack(alignment: .leading) {
                Text(title)
                    .padding([.top, .leading], 10)
                Text(content)
                    .foregroundColor(.gray)
                    .font(.system(size: copiable ? 12 : 18, design: .monospaced))
                    .padding([.leading, .bottom], 10)
                    .padding(.top, 1)
            }
            Spacer()
            if copiable {
                Button(action: {
                    UIPasteboard.general.string = content
                }) {
                    Image(systemName: "square.on.square")
                }
                .padding(10)
            }
        }
        .background(Color(UIColor.secondarySystemGroupedBackground))
        .cornerRadius(10)
    }
}

struct ContentCard_Previews: PreviewProvider {
    static var previews: some View {
        ContentCard(title: "title", content: "content", copiable: true)
    }
}
