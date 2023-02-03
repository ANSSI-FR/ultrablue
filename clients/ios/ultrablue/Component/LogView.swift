//
//  LogEntryView.swift
//  ultrablue
//
//  Created by loic buckwell on 18/07/2022.
//

import SwiftUI

struct LogView: View {
    @StateObject var log: Log
    var fontSize = 14.0
    
    var body: some View {
        HStack(alignment: .top, spacing: 0) {
            Text("[")
                .padding(.leading, 10)
                .font(.system(size: fontSize, design: .monospaced))
            switch (log.success) {
            case true:
                Text("OK")
                    .font(.system(size: fontSize, design: .monospaced))
                    .foregroundColor(Color(UIColor.systemGreen))
            case false:
                Text("KO")
                    .font(.system(size: fontSize, design: .monospaced))
                    .foregroundColor(Color(UIColor.systemRed))
            default:
                Text("  ")
                    .font(.system(size: fontSize, design: .monospaced))
            }
            Text("]")
                .font(.system(size: fontSize, design: .monospaced))
            Text(log.string)
                .padding(.leading, 10)
                .font(.system(size: fontSize, design: .monospaced))
            Spacer()
        }
        .frame(maxWidth: .infinity)
    }
}

func prepareLog(log: Log) -> Log {
    log.update(progress: 450)
    return log
}

struct LogView_Previews: PreviewProvider {
    
    static var previews: some View {
        LogView(log: prepareLog(log: Log("Fetching", tasksize: 50)))
    }
}
