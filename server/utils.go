// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/skip2/go-qrcode"
	"golang.org/x/term"
)

/*
	Seals the given key with the TPM Storage Root Key
	and stores it under two files named after the given
	uuid. If those files already exists, an error is
	returned.
*/
func storeKey(uuid string, key []byte) error {
	var priv, pub []byte
	var pin []byte
	var err error
	var fpriv, fpub *os.File

	if *withpin {
		fmt.Println("Choose a PIN to seal the encryption key on disk:")
		if pin, err = term.ReadPassword(syscall.Stdin); err != nil {
			return err
		}
	}
	if priv, pub, err = TPM2_Seal(key, string(pin)); err != nil {
		return err
	}
	if err = os.MkdirAll(ULTRABLUE_KEYS_PATH, os.ModeDir); err != nil {
		return err
	}
	if fpriv, err = os.OpenFile(ULTRABLUE_KEYS_PATH + uuid, os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0600); err != nil {
		return err
	}
	defer fpriv.Close()
	if fpub, err = os.OpenFile(ULTRABLUE_KEYS_PATH + uuid + ".pub", os.O_WRONLY | os.O_CREATE | os.O_EXCL, 0600); err != nil {
		return err
	}
	defer fpub.Close()
	if _, err = fpriv.Write(priv); err != nil {
		return err
	}
	if _, err = fpub.Write(pub); err != nil {
		return err
	}
	return nil
}

/*
	Gets the sealed key from the ultrablue keys directory
	and tries to unseal it with the TPM Storage Root Key.
	Returns the unsealed key on success
*/
func loadKey(uuid string) ([]byte, error) {
	var priv, pub, key []byte
	var pin []byte
	var err error

	if *withpin {
		fmt.Println("Please enter the PIN used to seal the encryption key:")
		if pin, err = term.ReadPassword(syscall.Stdin); err != nil {
			return nil, err
		}
	}
	if priv, err = os.ReadFile("/etc/ultrablue/" + uuid); err != nil {
		return nil, err
	}
	if pub, err = os.ReadFile("/etc/ultrablue/" + uuid + ".pub"); err != nil {
		return nil, err
	}
	if key, err = TPM2_Unseal(priv, pub, string(pin)); err != nil {
		return nil, err
	}
	return key, nil
}

/*
	generateQRCode generates a QR code containing the
	string given as parameter, and returns it in an
	ascii art string.
*/
func generateQRCode(data string) (string, error) {
	qr, err := qrcode.New(data, qrcode.Low)
	if err != nil {
		return "", err
	}
	return qr.ToSmallString(false), nil
}

/*
	Returns true if the data only contains zeros, false otherwise
*/
func onlyContainsZeros(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}
