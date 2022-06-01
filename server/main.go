// The ultrablue-server program starts the Ultrablue server side application.
package main

import (
	"context"
	"flag"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/go-ble/ble/linux/hci/evt"
)

// Command line arguments (global variables)
var (
	tpmPath  = flag.String("tpm", "/dev/tpm0", "Path of the TPM device to use")
	enroll   = flag.Bool("enroll", false, "Must be true for a first time attestation")
	loglevel = flag.Int("loglevel", 1, "Indicates the level of logging, between 0 (no logs) and 2 (verbose)")
)

func onConnection(evt evt.LEConnectionComplete) {
	log(1, "New device connection")
}

func onDisconnection(evt evt.DisconnectionComplete) {
	log(1, "Device disconnection")
}

func main() {
	flag.Parse()
	log(0, "Starting Ultrablue server")
	usage(*enroll)

	log(1, "Opening the default HCI device")
	device, err := linux.NewDevice(ble.OptConnectHandler(onConnection), ble.OptDisconnectHandler(onDisconnection))
	if err != nil {
		logErr(err)
		return
	}
	defer device.Stop()

	log(1, "Setting device as default one for future BLE communications")
	ble.SetDefaultDevice(device)

	if *enroll {

		log(1, "Starting device enrollment")
		/*
			TODO:
			1. Generate an AES key and an initial IV
			2. Create and display a QR code with MAC / AES key / IV in cbor format
			3. Store the AES key (and the IV) in the most secure way we can (to determine)
		*/
	}

	log(1, "Creating ultrablue service")
	ultrablueSvc := ble.NewService(ultrablueSvcUUID)

	log(1, "Adding characteristics to ultrablue service")
	if *enroll {
		ultrablueSvc.AddCharacteristic(registrationChr())
	}
	ultrablueSvc.AddCharacteristic(authenticationChr())
	ultrablueSvc.AddCharacteristic(credentialActivationChr())
	ultrablueSvc.AddCharacteristic(attestationChr())
	ultrablueSvc.AddCharacteristic(responseChr())

	log(1, "Binding ultrablue service to BLE HCI")
	if err = ble.AddService(ultrablueSvc); err != nil {
		logErr(err)
		return
	}

	log(1, "Start advertising ultrablue service and it's characteristics")
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	if err = ble.AdvertiseNameAndServices(ctx, "Ultrablue server", ultrablueSvc.UUID); err != nil {
		logErr(err)
		return
	}
}
