// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

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

// ultrablue-server
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/go-ble/ble/linux/hci/evt"
	"github.com/sirupsen/logrus"
)

// Command line arguments (global variables)
var (
	tpmPath  = flag.String("tpm", "/dev/tpm0", "Path of the TPM device to use")
	enroll   = flag.Bool("enroll", false, "Must be true for a first time attestation")
	loglevel = flag.Int("loglevel", 0, "Indicates the level of logging, 0 is the minimum, 2 is the maximum")
)

func initLogger(loglevel int) {
	switch loglevel {
	case 1:
		logrus.SetLevel(logrus.InfoLevel)
	case 2:
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.ErrorLevel)
	}
}

func onConnection(evt evt.LEConnectionComplete) {
	logrus.Info("New device connection")
}

func onDisconnection(evt evt.DisconnectionComplete) {
	logrus.Info("Device disconnection")
}

func main() {
	flag.Parse()
	usage(*enroll)
	initLogger(*loglevel)

	logrus.Info("Opening the default HCI device")
	device, err := linux.NewDevice(ble.OptConnectHandler(onConnection), ble.OptDisconnectHandler(onDisconnection))
	if err != nil {
		logrus.Fatal(err)
	}
	ble.SetDefaultDevice(device)
	device.SetCentralRole()
	defer device.Stop()

	if *enroll {
		logrus.Info("Generating registration QR code")

		rbytes, err := getTPMRandom(32)
		if err != nil {
			logrus.Fatal(err)
		}
		key, iv := rbytes[0:16], rbytes[16:32]

		mac := device.Address().String()

		qrcode, err := generateRegistrationQR(key, iv, mac)
		if err != nil {
			logrus.Fatal(err)
		}
		fmt.Println(qrcode)
	}

	logrus.Info("Registering ultrablue service and characteristics")
	ultrablueSvc := ble.NewService(ultrablueSvcUUID)

	errc := make(chan error)
	rspc := make(chan int)

	if *enroll {
		ultrablueSvc.AddCharacteristic(registrationChr(errc))
	}
	ultrablueSvc.AddCharacteristic(authenticationChr(errc))
	ultrablueSvc.AddCharacteristic(credentialActivationChr(errc))
	ultrablueSvc.AddCharacteristic(attestationChr(errc))
	ultrablueSvc.AddCharacteristic(responseChr(errc, rspc))

	if err = ble.AddService(ultrablueSvc); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Start advertising")
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	go ble.AdvertiseNameAndServices(ctx, "Ultrablue server", ultrablueSvc.UUID)

	select {
	case <-ctx.Done():
		logrus.Fatal(ctx.Err())
	case err := <-errc:
		logrus.Fatal(err)
	case <-rspc:
		logrus.Info("Attestation succeeded")
	}
}
