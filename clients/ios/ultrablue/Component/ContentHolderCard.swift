//
//  ContentHolderCard.swift
//  ultrablue
//
//  Created by loic buckwell on 19/08/2022.
//

import SwiftUI

struct ContentHolderCard<Content: View>: View {
    var title: String
    var content: () -> Content
    var details: String?
    
    init(title: String, details: String? = nil, @ViewBuilder content: @escaping () -> Content) {
        self.title = title
        self.details = details
        self.content = content
    }
    
    var body: some View {
        HStack(alignment: .top) {
            VStack(alignment: .leading) {
                Text(title + ":")
                    .bold()
                    .padding(.leading, 10)
                HStack(alignment: .top) {
                    VStack(content: content)
                    Spacer()
                }
                .padding([.leading, .vertical], 10)
                .background(Color(UIColor.secondarySystemGroupedBackground))
                .cornerRadius(10)
                if let d = details {
                    Text(d)
                        .padding(.leading, 10)
                        .foregroundColor(.gray)
                        .font(.system(size: 14))
                }
            }
            .padding(10)
        }
    }
}

struct ContentHolderCard_Previews: PreviewProvider {
    static var previews: some View {
        ContentHolderCard(title: "Modified PCRs") {
            Text("Heyho")
        }
    }
}
