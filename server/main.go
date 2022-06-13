// The ultrablue-server program starts the Ultrablue server side application.
package main

import (
	"context"
	"flag"
	"fmt"

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
	ble.SetDefaultDevice(device)
	defer device.Stop()

	if *enroll {
		log(1, "Generating registration QR code")

		rbytes, err := getTPMRandom(32)
		if err != nil {
			logErr(err)
			return
		}
		key, iv := rbytes[0:16], rbytes[16:32]

		mac := device.Address().String()
		log(2, "HCI's MAC address:", mac)

		qrcode, err := generateRegistrationQR(key, iv, mac)
		if err != nil {
			logErr(err)
			return
		}
		fmt.Println(qrcode)
	}

	log(1, "Registering ultrablue service")
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
		logErr(err)
		return
	}

	log(1, "Start advertising")
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))
	go ble.AdvertiseNameAndServices(ctx, "Ultrablue server", ultrablueSvc.UUID)

	select {
	case <-ctx.Done():
		logErr(ctx.Err())
	case err := <-errc:
		logErr(err)
	case <-rspc:
		log(1, "Attestation succeeded")
	}
}
