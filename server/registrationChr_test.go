// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestRegistrationChr(t *testing.T) {
	var errc chan error
	chr := registrationChr(errc)

	if chr.UUID.String() != registrationChrUUID.String() {
		t.Logf("Wrong UUID for registrationChr: got %s, expected %s", chr.UUID.String(), registrationChrUUID.String())
		t.Fail()
	}
}
