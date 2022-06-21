// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/go-ble/ble"
	"github.com/sirupsen/logrus"
)

/*
	authenticationChr authenticates a remote device for an attestation.
	Apart from an enrollment, this will be the first characteristic exposed, and
	while the verifier is not authenticated, the only available one.

	When authenticated, the other characteristics will be made available, and all
	the messages will be encrypted with AES GCM
*/
func authenticationChr(errc chan error) *ble.Characteristic {

	logrus.Debug("Creating new characteristic: authenticationChr")
	chr := ble.NewCharacteristic(authenticationChrUUID)

	// TODO: HandleRead - Generate an IV and a nonce, and send the IV in clear text, plus the nonce,
	// encrypted with the shared key and the IV

	// TODO: HandleWrite - Get the decrypted nonce from the verifier, and check it against the generated one.
	// The authentication is approved if they match

	return chr
}
