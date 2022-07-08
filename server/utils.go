// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
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
	generateQRCode generates a QR code containing the
	string given as parameter, and returns it in an
	ascii art string.
*/
func generateQRCode(data string) (string, error) {
	qr, err := qrcode.New(fmt.Sprintf("%s\n", data), qrcode.Low)
	if err != nil {
		return "", err
	}
	return qr.ToString(false), nil
}
