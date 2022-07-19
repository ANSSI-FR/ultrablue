//
//  Log.swift
//  ultrablue
//
//  Created by loic buckwell on 18/07/2022.
//

import Foundation

extension Log {
    enum Kind {
        case standard, progress
    }
}

class Log {
    private let uuid = UUID()
    private let kind: Kind
    private let value: String
    
    var success: Bool?
    var state: (progress: UInt, total: UInt)?

    var string: String {
        get {
            var val = value
            if kind == .progress {
                let progressBar = drawProgressBar(state!)
                val += "\n" + progressBar
            }
            return val
        }
    }

    init(_ msg: String) {
        self.kind = .standard
        self.value = msg
    }
    
    init(_ msg: String, tasksize: UInt) {
        self.kind = .progress
        self.state = (0, tasksize)
        self.value = msg
    }

    func complete(success: Bool) {
        self.success = success
    }
    
    func update(progress: UInt) {
        if kind == .progress {
            if progress >= state!.total {
                success = true
                state!.progress = state!.total
            } else {
                state!.progress = progress
            }
        }
    }
    
    private func drawProgressBar(_ state: (progress: UInt, total: UInt)) -> String {
        let barLength = 20
        let percent = Float(state.progress) / Float(state.total)
        let filled = Int(percent * (Float(barLength) - 2))
        let empty = barLength - (2 + filled)

        let barBody = String(repeating: "=", count: filled)
        return ("[" + barBody.replaceLast(with: ">") + String(repeating: " ", count: empty) + "] \(state.progress)/\(state.total)")
    }
    
}

class Logger {
    var logs: [Log]
    
    init() {
        self.logs = [Log]()
    }

    func push(log: Log) {
        self.logs.append(log)
    }
}
