// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

// ultrablue-server
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/sirupsen/logrus"
)

// Command line arguments - Global variables
var (
	enroll       = flag.Bool("enroll", false, "Must be set for a first time attestation (known as the enrollment)")
	loglevel     = flag.Int("loglevel", 1, "Indicates the level of logging, 0 is the minimum, 3 is the maximum")
	mtu          = flag.Int("mtu", 500, "Set a custom MTU, which is basically the max size of the BLE packets")
	pcrextend    = flag.Bool("pcr-extend", false, "Extend the 9th PCR with the verifier secret on attestation success")
	withpin      = flag.Bool("with-pin", false, "Use a PIN to seal the encryption key to the TPM (default is sealing to the SRK without password)")
)

// Encryption key used at enroll time. It needs to be globally available
// to be accessible from the protocol functions
var enrollkey []byte

const ULTRABLUE_KEYS_PATH = "/etc/ultrablue/"

/*
	initLogger sets the level of logging
	according to the loglevel parameter.
	Here is a short description of the log
	levels:
		- 0: No log
		- 1: Protocol steps logs
		- 2: Debug messages
		- 3: BLE packets trace
*/
func initLogger(loglevel int) {
	switch loglevel {
	case 1:
		logrus.SetLevel(logrus.InfoLevel)
	case 2:
		logrus.SetLevel(logrus.DebugLevel)
	case 3:
		logrus.SetLevel(logrus.TraceLevel)
	default:
		logrus.SetLevel(logrus.ErrorLevel)
	}
}

/*
	ARCHITECTURE OVERVIEW

	Ultrablue is a client-server application that operates over
	Bluetooth Low Energy (BLE). This tool acts as the server.

	BLE is a client-driven transport layer, in the sense that
	the server exposes characteristics and it is up to the client to read or
	write them whenever it wants.

	Ultrablue implements an inversion of control through several components,
	abstracting the BLE layer and allowing the server to drive the exchange.
	Each of those components, briefly described here, is implemented in a
	dedicated file named after the functionality.

	- The characteristic: Bluetooth Low Energy sends/receives data through
	  characteristics. When a characteristic is advertised by a device, clients
	  are able to see it and can read/write on it if allowed by the advertising
	  device. When a client performs a read/write operation, the characteristic
	  will process it through handlers.
	  Ultrablue only exposes one characteristic, enabled both for reading and
	  writing.

	- The state: The state is a data structure that holds information about the
	  currently running read/write operation on the characteristic for a
	  connection.
	  It has a go channel, that abstracts the characteristic handlers and makes
	  possible to read/write full messages to the channel without dealing with BLE
	  internals (chunking, size prefix...).
	  The state is created on the first client interaction with the characteristic,
	  and lives as long as the client connection does.
	  When the state is created, it also runs a protocol instance that operates
	  on the above-mentioned channel and runs in a dedicated goroutine.

	- The session: It wraps the state channel and keeps information on how to
	  read/write data into it, e.g. if messages must be encrypted/decrypted.
	  A StartEncryption method is available so that the caller can make the session
	  encrypted whenever they want.
	  The session also exposes two functions to communicate with clients: SendMsg
	  and recvMsg. These functions make use of the abstractions provided by lower
	  layers to operate on the characteristic (through the go channel of the
	  state); upper layers should never need to call the lower layers directly.

	- The protocol: It implements the actual remote attestation routine. It is
	  split in several steps, each implemented in its own function. Thanks to
	  previous abstractions, the server doesn't care of the transport layer and
	  only relies on the session and its exported methods/functions to communicate
	  with clients.

	Note: The server only accepts one simulteanous client.
*/

func main() {
	flag.Parse()
	initLogger(*loglevel)

	logrus.Info("Opening the default HCI device")
	device, err := linux.NewDevice()
	if err != nil {
		logrus.Fatal(err)
	}
	ble.SetDefaultDevice(device)
	defer device.Stop()

	if *enroll {
		logrus.Info("Generating symmetric key")
		if enrollkey, err = TPM2_GetRandom(32); err != nil {
			logrus.Fatal(err)
		}
	}

	logrus.Info("Registering ultrablue service and characteristic")
	ultrablueSvc := ble.NewService(ultrablueSvcUUID)
	ultrablueSvc.AddCharacteristic(UltrablueChr(*mtu))
	if err := ble.AddService(ultrablueSvc); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Start advertising")
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	go ble.AdvertiseNameAndServices(ctx, "Ultrablue server", ultrablueSvc.UUID)

	if *enroll {
		logrus.Info("Generating enrollment QR code")
		addr := device.Address().String()
		json := fmt.Sprintf(`{"addr":"%s","key":"%x"}`, addr, enrollkey)
		qrcode, err := generateQRCode(json)
		if err != nil {
			logrus.Fatal(err)
		}
		fmt.Print(qrcode)
	}

	select {
	case <-ctx.Done():
		logrus.Fatal(ctx.Err())
	}
}
