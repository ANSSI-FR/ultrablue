// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
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

func extendPCR(index int, secret []byte) error {
	rwc, err := tpm2.OpenTPM()
	if err != nil {
		return err
	}
	defer rwc.Close()

	// TODO: The following extends the PCR, but does not adds
	// any event log entry. This is sufficient for our needs,
	// but it means that after an ultrablue attestation, the
	// eventlog will differ from the pcrs, thus making any
	// replay fail.
	// NOTE: It seems that it's not possible to add an event log
	// entry directly from the tss stack, and that we need to use
	// exposed UEFI function pointers.
	err = tpm2.PCREvent(rwc, tpmutil.Handle(index), secret)
	if err != nil {
		return err
	}
	return nil
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
