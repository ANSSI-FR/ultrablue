//
//  Log.swift
//  ultrablue
//
//  Created by loic buckwell on 18/07/2022.
//

import Foundation
import SwiftUI
import Combine

extension Log {
    enum Kind {
        case standard, progress
    }
}

class Log: Identifiable, ObservableObject {
    private let uuid = UUID()
    private let kind: Kind
    private let value: String
    
    @Published var success: Bool?
    @Published var state: (progress: UInt, total: UInt)?

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

    init(_ msg: String, success: Bool? = nil) {
        self.kind = .standard
        self.value = msg
        self.success = success
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
    
    func update(delta: UInt) {
        if kind == .progress {
            if state!.progress + delta >= state!.total {
                success = true
                state!.progress = state!.total
            } else {
                state!.progress += delta
            }
        }
    }
    
    private func drawProgressBar(_ state: (progress: UInt, total: UInt)) -> String {
        let barLength = 23
        let percent = Float(state.progress) / Float(state.total)
        let filled = Int(percent * (Float(barLength) - 2))
        let empty = barLength - (2 + filled)

        let barBody = String(repeating: "=", count: filled)
        return ("[" + barBody.replaceLast(with: ">") + String(repeating: " ", count: empty) + "] \(state.progress)/\(state.total)")
    }
    
}

class Logger: ObservableObject {
    @Published private var logs: [Log]
    private var onFailure: () -> Void
    
    var lines: [Log] {
        return self.logs
    }
    
    init(onFailure: @escaping () -> Void = {}) {
        self.onFailure = onFailure
        self.logs = [Log]()
    }

    func push(log: Log) {
        if !logs.isEmpty && !isLastComplete() {
            completeLast(success: false)
        }
        self.logs.append(log)
        if log.success == false {
            onFailure()
            self.setOnFailureCallback ({})
        }
    }
    
    func updateLast(progress: UInt) {
        self.logs.last?.update(progress: progress)
    }
    
    func updateLast(delta: UInt) {
        self.logs.last?.update(delta: delta)
    }
    
    func completeLast(success: Bool) {
        self.logs.last?.complete(success: success)
        if success == false {
            onFailure()
            self.setOnFailureCallback ({})
        }
    }
    
    func isLastComplete() -> Bool {
        return self.logs.last?.success != nil
    }
    
    func clear() {
        self.logs.removeAll()
    }
    
    func debug() {
        print("=== DEBUG ===")
        for log in logs {
            print("log: \(log.string), success: \(log.success ?? false)")
        }
    }
    
    func setOnFailureCallback(_ callback: @escaping () -> Void) {
        self.onFailure = callback
    }
}
