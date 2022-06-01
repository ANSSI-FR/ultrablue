package main

import "testing"

func TestRegistrationChr(t *testing.T) {
	chr := registrationChr()

	if chr.UUID.String() != registrationChrUUID.String() {
		t.Logf("Wrong UUID for registrationChr: got %s, expected %s", chr.UUID.String(), registrationChrUUID.String())
		t.Fail()
	}
}
