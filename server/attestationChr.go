// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/go-ble/ble"
)

/*
	attestationChr provides all the data the verifier needs
	to perform the remote attestation, including the eventlog and the quotes
*/
func attestationChr(errc chan error) *ble.Characteristic {

	log(2, "attestationChr")
	chr := ble.NewCharacteristic(attestationChrUUID)

	// TODO: HandleWrite - Get a nonce generated on the verifier side, to avoid
	// replay attacks

	// TODO: HandleRead - Send back the attestation data, the quotes will include
	// the nonce (and is signed with the attestation key)

	return chr
}
