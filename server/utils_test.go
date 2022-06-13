package main

import "testing"

func TestGenerateRegistrationQR(t *testing.T) {
	var k = []byte("imakey")
	var i = []byte("imaniv")
	var m = "imanaddr"

	q, err := generateRegistrationQR(k, i, m)
	if err != nil {
		t.Log("An error occured:", err)
		t.Fail()
	}

	if len(q) == 0 {
		t.Log("Got empty string for qrcode")
		t.Fail()
	}
}

// TODO: move to integration tests. This test currently
// needs to run as root, in order to access the TPM
func TestGetTPMRandom(t *testing.T) {
	for i := 0; i < 50; i++ {
		rb, err := getTPMRandom(uint16(i * 3))
		if err != nil {
			t.Log(err)
			t.Fail()
		} else if len(rb) != i*3 {
			t.Logf("Error: expected to get %d bytes, got %d", i*3, len(rb))
		}
	}
}
