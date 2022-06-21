// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/go-ble/ble"
	"github.com/sirupsen/logrus"
)

/*
	responseChr gets an optional secret from the verifier if
	the attestation succeeds.
	If a secret is sent back it is then used to extend a PCR.
*/
func responseChr(errc chan error, rspc chan int) *ble.Characteristic {

	logrus.Debug("Creating new characteristic: responseChr")
	chr := ble.NewCharacteristic(responseChrUUID)

	// TODO: HandleWrite - Read the attestation response and extend the 9th PCR

	// TODO: HandleRead (idea for later) - Available only during enrollment, allow to specify
	// and send the secret that the verifier will send back in case of a successful attestation
	// with e.g. a --secret option

	return chr
}
