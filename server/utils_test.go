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
