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
	loglevel = flag.Int("loglevel", 0, "Indicates the level of logging, 0 is the minimum, 3 is the maximum")
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
   usage prints informations to guide the user at the
   start of the program.
*/
func usage(erl bool) {
	var yellow = "\033[33m"
	var reset = "\033[0m"

	if erl {
		fmt.Println(yellow + `To enroll a new verifier device:

   1. Install the Ultrablue application on your smartphone
   2. From it, push the + button on the top-right corner
   3. Scan the following QR code
` + reset)
	} else {
		fmt.Println(yellow + `To perform the attestation:

   1. Open the Ultrablue application on an enrolled device
   2. Find this computer in the list of known attesters
   3. Tap the play button
` + reset)
	}
}

/*
	ARCHITECTURE

	Ultrablue is a client-server application, that operates over
	Bluetooth Low Energy. This tool acts as the server.
	The client has the full control over the conversation, thus we
	implemented some sort of control inversion, with go-routines,
	channels and objects as follow:

		- A characteristic is exposed and will handle the read/write
		operations of the client and the chunking of the packets. It will
		read/write on a message channel once an operation is complete.
		The goal of the characteristic is to abstract the BLE transport layer.

		- A protocol function that will implement the protocol logic. It will
		be started as a go-routine, each time a new client read or write on the
		characteristic for the first time.

		- A state structure that contains everything the characteristic and the
		protocol routines needs to know about the client connection to work
		together.

	NOTES:

		- The server only accepts one simulteanous client.
		- The protocol diagram can be found in the README.md file.
*/

func main() {
	flag.Parse()
	usage(*enroll)
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
