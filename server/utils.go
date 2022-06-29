// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

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
