// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/go-ble/ble"
)

/*
	credentialActivationChr performs the credential activation, which consists in:
	- Generating and sending an attestation key (AK) to the verifier
	- Proving him than it comes from the attester's TPM he trusts
*/
func credentialActivationChr(errc chan error) *ble.Characteristic {

	log(2, "credActivationChr")
	chr := ble.NewCharacteristic(credentialActivationChrUUID)

	// TODO: HandleRead (1) - Generate and send an attestation key

	// TODO: HandleWrite - Get the credential blob generated on the verifier side
	// with the make_credential primitive

	// TODO: HandleRead (2) - decrypt the credential blob with activate_credential
	// and send back the original nonce to the verifier

	return chr
}
