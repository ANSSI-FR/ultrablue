//
//  BLE.swift
//  ultrablue
//
//  Created by loic buckwell on 21/07/2022.
//

import Foundation
import CoreBluetooth
import Combine
import SwiftUI

let ultrablueSvcUUID = CBUUID(string: "ebee1789-50b3-4943-8396-16c0b7231cad")
let ultrablueChrUUID = CBUUID(string: "ebee1790-50b3-4943-8396-16c0b7231cad")

class BLEManager: NSObject {
    private var manager: CBCentralManager!
    private var logger: Logger? = nil
    private var attester: CBPeripheral? = nil
    private var attestationChr: CBCharacteristic? = nil
    private var onAttesterFound: (() -> Void)? = nil
    private var onMsgReadCallback: ((Data) -> Void)? = nil
    private var onMsgWriteCallback: (() -> Void)? = nil
    
    private var shutdownLock: Bool = false
    private var shutdownErr: Bool = false
    private var message: Data = Data()
    private var messageLength: UInt = 0
    private var messageTag: String = ""
    
    override init() {
        super.init()
        manager = CBCentralManager(delegate: self, queue: .main)
    }
    
    func setLogger(logger: Logger) {
        self.logger = logger
    }
    
    func searchForAttestingDevice(onAttesterFound: @escaping () -> Void) {
        if manager.state == .poweredOn {
            reset()
            self.onAttesterFound = onAttesterFound
            logger?.push(log: Log("Scanning for attesting device"))
            let options: [String: Any] = [CBCentralManagerScanOptionAllowDuplicatesKey: NSNumber(value: false)]
            manager.scanForPeripherals(withServices: nil, options: options)
        }
    }
    
    func shutdown(err: Bool = false) {
        if shutdownLock {
            return
        }
        shutdownLock = true
        logger?.push(log: Log("Shutting down connection"))
        if manager.isScanning {
            manager.stopScan()
        }
        if let at = attester {
            attestationChr = nil // To make sure the disconnection will not block
            shutdownErr = err
            self.manager.cancelPeripheralConnection(at)
        } else {
            logger?.completeLast(success: !err)
        }
    }
    
    func reset() {
        attester = nil
        attestationChr = nil
        message = Data()
        messageLength = 0
        messageTag = ""
        shutdownErr = false
        shutdownLock = false
    }
    
    func sendMsg(msg: Data) {
        if let chr = attestationChr {
            // TODO: Check that the message + prepended size is shorter than the MTU as we don't handle chunk for writing operations
            logger?.push(log: Log("Sending " + messageTag, tasksize: UInt(msg.count)))
            let size = UInt(msg.count).toLittleEndianData()
            attester?.writeValue(size + msg, for: chr, type: .withResponse)
        }
    }
    
    func recvMsg() {
        if let chr = attestationChr {
            attester?.readValue(for: chr)
        }
    }
    
    func setOnMsgReadCallback(callback: @escaping (Data) -> Void) {
        self.onMsgReadCallback = callback
    }
    
    func setOnMsgWriteCallback(callback: @escaping () -> Void) {
        self.onMsgWriteCallback = callback
    }
    
    func getAttesterUUID() -> UUID? {
        if let a = attester {
            return a.identifier
        } else {
            return nil
        }
    }

}

extension BLEManager: CBCentralManagerDelegate {
   
    func centralManagerDidUpdateState(_ central: CBCentralManager) {
        switch central.state {
        case .poweredOff:
            print("BLE is powered off")
        case .poweredOn:
            print("BLE is powered On")
        default:
            print("default")
        }
    }
    
    func centralManager(_ central: CBCentralManager, didDiscover peripheral: CBPeripheral, advertisementData: [String : Any], rssi RSSI: NSNumber) {
        // TODO: Scan for server with the name "ultrablue-PIN", where pin is a pin in the qrcode, or scan for a known device (with uuid)
        let advertisedName = advertisementData["kCBAdvDataLocalName"] as? String?
        if advertisedName == "ultrablue-server" {
            central.stopScan()
            logger?.completeLast(success: true)
            logger?.push(log: Log("Device found, connecting"))
            attester = peripheral
            central.connect(peripheral)
        }
    }
    
    func centralManager(_ central: CBCentralManager, didConnect peripheral: CBPeripheral) {
        logger?.completeLast(success: true)
        logger?.push(log: Log("Searching for Ultrablue service"))
        peripheral.delegate = self
        peripheral.discoverServices(nil)
    }
    
    func centralManager(_ central: CBCentralManager, didDisconnectPeripheral peripheral: CBPeripheral, error: Error?) {
        if !shutdownLock {
            shutdownLock = true
            logger?.push(log: Log("Device just disconnected", success: false))
        } else {
            logger?.completeLast(success: !shutdownErr)
        }
    }
    
    func setMessageTag(_ tag: String) {
        self.messageTag = tag
    }

}

extension BLEManager: CBPeripheralDelegate {
    
    func peripheral(_ peripheral: CBPeripheral, didDiscoverServices error: Error?) {
        if let services = peripheral.services {
            for service in services {
                if service.uuid == ultrablueSvcUUID {
                    logger?.completeLast(success: true)
                    logger?.push(log: Log("Searching for Ultrablue characteristic"))
                    peripheral.discoverCharacteristics(nil, for: service)
                    return
                }
            }
        }
        logger?.completeLast(success: false)
    }
    
    func peripheral(_ peripheral: CBPeripheral, didDiscoverCharacteristicsFor service: CBService, error: Error?) {
        if let characteristics = service.characteristics {
            for characteristic in characteristics {
                if characteristic.uuid == ultrablueChrUUID {
                    self.attestationChr = characteristic
                    logger?.completeLast(success: true)
                    onAttesterFound?()
                    return
                }
            }
        }
        logger?.completeLast(success: false)
    }
    
    func peripheral(_ peripheral: CBPeripheral, didWriteValueFor characteristic: CBCharacteristic, error: Error?) {
        // Error checking
        if error != nil {
            logger?.completeLast(success: false)
            return
        }
        
        // We don't handle chunking, so return as soon as we wrote a message. We can do that because the client isn't supposed to send messages longer than MTU.
        logger?.fillLast()
        self.onMsgWriteCallback?()
    }
    
    func peripheral(_ peripheral: CBPeripheral, didUpdateValueFor characteristic: CBCharacteristic, error: Error?) {
        // Error checking
        if error != nil || characteristic.value == nil {
            logger?.completeLast(success: false)
            return
        }
        let data = characteristic.value!
        
        // Check if it's the first messages's packet.
        // If it is, take the 4 first bytes as the message size, little endian encoded
        if messageLength == 0 {
            messageLength = data.subdata(in: 0..<4).toUInt()
            message = data.subdata(in: 4..<data.count)
            logger?.push(log: Log("Fetching " + messageTag, tasksize: messageLength))
        } else {
            message += data
        }
        logger?.updateLast(delta: UInt(data.count))
        
        // Check if the whole message has been read
        if message.count < messageLength {
            self.recvMsg()
        } else {
            let completeMessage = message
            messageLength = 0
            message.removeAll()
            self.onMsgReadCallback?(completeMessage)
        }
    }
    
}
