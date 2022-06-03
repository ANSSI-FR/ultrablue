package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

/*
	generateAES128KeyWithIV generates and returns an AES 128 bits key,
	and a 128 bits IV.
*/
func generateAES128KeyWithIV() ([]byte, []byte, error) {
	key := make([]byte, 32)
	iv := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return nil, nil, err
	}
	_, err = rand.Read(iv)
	if err != nil {
		return nil, nil, err
	}
	return key, iv, nil
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
