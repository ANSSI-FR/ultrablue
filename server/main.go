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

// Command line arguments (global variables)
var (
	enroll   = flag.Bool("enroll", false, "Must be true for a first time attestation")
	loglevel = flag.Int("loglevel", 1, "Indicates the level of logging, 0 is the minimum, 3 is the maximum")
	mtu      = flag.Int("mtu", 500, "Set a custom MTU, which is basically the max size of the BLE packets")
)

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
	ARCHITECTURE

	Ultrablue is a client-server application, that operates over
	Bluetooth Low Energy. This tool acts as the server.
	Each step of the protocol is implemented in a characteristic,
	and the client must read/write on those characteristic
	successively to perform the remote attestation.
	The protocol diagram can be found in the README.md file.

	Characteristics implementation details:

		- Each characteristic is declared in its own file, ending with Chr.go.

		- As the chunking of the packets is not handled by the ble package,
		each characteristic maintains its state in the context associated
		with the connection.
		It's up to the client to read/write enough times to complete the
		operation, the server can't drive the communicatin and send every
		chunks in a for loop.
		TODO: the server should disconnect the client if it does not follow the expected steps of the protocol

		- The r/w handlers of the characteristics runs in a goroutine, thus
		errors/success are transmitted to the main routine through channels.

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
		qrcode, err := generateQRCode(addr)
		if err != nil {
			logrus.Fatal(err)
		}
		fmt.Println(qrcode)
	}

	select {
	case <-ctx.Done():
		logrus.Fatal(ctx.Err())
	}
}
