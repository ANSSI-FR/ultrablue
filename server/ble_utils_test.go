package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/go-ble/ble"
)

func TestSendBLEPacket_OneShot(t *testing.T) {
	var cases = []struct {
		offset   int
		msglen   int
		mtu      int
		expected bool
		name     string
	}{
		{0, 0, 20, true, "Null sized message"},
		{0, 42, 20, false, "Message length > MTU"},
		{0, 30, 45, true, "Message length < MTU"},
		{0, 42, 42, false, "Message length == MTU"},
		{0, 42, 44, false, "Message length + prefix > MTU"},
		{20, 41, 20, false, "Message length - offset > MTU"},
		{42, 50, 20, true, "Message length - offset < MTU"},
		{20, 40, 20, true, "Message length - offset == MTU"},
	}

	for _, c := range cases {
		var responseBuffer = bytes.NewBuffer(make([]byte, 0, 10000))
		var rw = ble.NewResponseWriter(responseBuffer)

		var msg = make([]byte, c.msglen)
		rand.Read(msg)

		result, err := sendBLEPacket(&c.offset, msg, c.mtu, rw)
		if err != nil {
			t.Errorf("[%s]: %s", c.name, err)
		}
		if result != c.expected {
			t.Errorf("[%s]: expected: %t, got: %t (msg length: %d, offset: %d, MTU: %d)", c.name, c.expected, result, c.msglen, c.offset, c.mtu)
		}
	}
}

func TestSendBLEPacket_RealCases(t *testing.T) {
	var cases = []struct {
		msglen   int
		mtu      int
		expected int
		name     string
	}{
		{0, 20, 1, "Null sized message"},
		{12, 20, 1, "Message length < MTU"},
		{20, 20, 2, "Message length == MTU"},
		{80, 20, 5, "Message length == MTU * 4"},
	}

	for _, c := range cases {
		var responseBuffer = bytes.NewBuffer(make([]byte, 0, 10000))
		var rw = ble.NewResponseWriter(responseBuffer)

		var msg = make([]byte, c.msglen)
		var offset, result int
		rand.Read(msg)

		for true {
			complete, err := sendBLEPacket(&offset, msg, c.mtu, rw)
			result += 1
			if err != nil {
				t.Errorf("[%s]: %s", c.name, err)
			}
			if complete {
				break
			}
		}

		if result != c.expected {
			t.Errorf("[%s]: expected exactly %d calls to send %d bytes with mtu %d, got %d", c.name, c.expected, c.msglen, c.mtu, result)
		}
		if c.msglen > 0 {
			if bytes.Compare(responseBuffer.Bytes()[4:], msg) != 0 {
				t.Errorf("[%s]: sended bytes differs from the original message:\nexpected: %x\ngot:      %x", c.name, msg, responseBuffer.Bytes()[4:])
			}
			prefix := int(binary.LittleEndian.Uint32(responseBuffer.Bytes()[:4]))
			if prefix != c.msglen {
				t.Errorf("[%s]: message is %d bytes long, but the prefix indicates %d", c.name, c.msglen, prefix)
			}
		}
	}
}
