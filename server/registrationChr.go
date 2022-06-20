// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/go-ble/ble"
)

/*
	registrationChr sends the EkPub and the EkCert to an enrolling
	device, to complete it's registration phase.
	This characteristic will only be registered if the enroll flag is set.
*/
func registrationChr(cerr chan error) *ble.Characteristic {

	log(2, "registrationChr")
	chr := ble.NewCharacteristic(registrationChrUUID)

	// TODO: HandleRead - Send the EkPub, and the EkCert

	// TODO: HandleWrite - Get verifier registration confirmation.
	// At this stage, we also save the registration data on attester side, if an error
	// occurs later, we'll be able to run a classic attestation to finish the enrollment.

	return chr
}
