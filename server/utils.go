package main

import (
	"encoding/base64"
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/skip2/go-qrcode"
)

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
	log(2, "encoded AES key:", ek)
	ei := base64.StdEncoding.EncodeToString(iv)
	log(2, "encoded IV:", ei)

	qr, err := qrcode.New(fmt.Sprintf("%s\n%s\n%s\n", ek, ei, mac), qrcode.Low)
	if err != nil {
		return "", err
	}

	return qr.ToString(false), nil
}
