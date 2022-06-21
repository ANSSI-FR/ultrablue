// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/base64"
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/skip2/go-qrcode"
)

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
	getTPMRandom gets @size random bytes from the TPM.
	This function is used to generate AES keys and IVs.
*/
func getTPMRandom(size uint16) ([]byte, error) {
	rwc, err := tpm2.OpenTPM()
	if err != nil {
		return nil, err
	}
	defer rwc.Close()

	rbytes, err := tpm2.GetRandom(rwc, size)
	if err != nil {
		return nil, err
	}
	return rbytes, nil
}

/*
	generateRegistrationQR generates the QRcode the remote device needs to scan
	to enroll itself. In the qrcode, there's:
	- The AES key, base64 encoded
	- The IV, base64 encoded
	- The MAC address
*/
func generateRegistrationQR(key, iv []byte, mac string) (string, error) {

	ek := base64.StdEncoding.EncodeToString(key)
	ei := base64.StdEncoding.EncodeToString(iv)

	qr, err := qrcode.New(fmt.Sprintf("%s\n%s\n%s\n", ek, ei, mac), qrcode.Low)
	if err != nil {
		return "", err
	}

	return qr.ToString(false), nil
}
