// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
	"time"

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

func TestRecvBLEPacket_OneShot(t *testing.T) {
	var cases = []struct {
		readlen  int
		preset   bool
		msglen   int
		expected bool
		desc     string
	}{
		{20, true, 20, true, "preset readlen - small mtu"},
		{500, true, 500, true, "preset readlen - medium mtu"},
		{1000, true, 1000, true, "preset readlen - big mtu"},
		{35, true, 50, false, "preset readlen - msg length > msglen"},
		{50, true, 35, false, "preset readlen - msg length < msglen"},

		{20, false, 20, true, "no readlen - small mtu"},
		{500, false, 500, true, "no readlen - medium mtu"},
		{1000, false, 1000, true, "no readlen - big mtu"},
		{35, false, 50, false, "no readlen - msg length > msglen"},
		{50, false, 35, false, "preset readlen - msg length < msglen"},
		{123431243, false, 23, false, "no readlen - msg length < msglen"},
		{0, false, 30, false, "no readlen - forgot readlen header"},
	}

	for _, c := range cases {
		var msg = make([]byte, c.msglen)
		if c.preset == false {
			header := make([]byte, 4)
			binary.LittleEndian.PutUint32(header[:], uint32(c.readlen))
			msg = append(header, msg...)
			c.readlen = 0
		}
		var req = ble.NewRequest(nil, msg, 0)
		var buf = make([]byte, 0)

		result := recvBLEPacket(&buf, &c.readlen, req)
		if result != c.expected {
			t.Errorf("[%s] - read: %d, msglen: %d | expected: %t, got %t", c.desc, c.readlen, c.msglen, c.expected, result)
		}
		if len(buf) != c.msglen {
			t.Errorf("[%s] - expected msg of length %d, got %d", c.desc, c.msglen, len(buf))
		}
	}
}

func TestRecvBLEPacket_With_SendBLEPacket(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		// Set a random MTU, the sendBLEPacket function should fix it if invalid
		var mtu = rand.Intn(500-20) + 20

		// Create a random message of a random size
		var msglen = rand.Intn(10000)
		var msg = make([]byte, msglen)
		rand.Read(msg)

		// Create the buffer to read the message in
		var out []byte

		// Function arguments, return values and helper variables
		var completewr, completerd bool
		var off, ml, counter int

		// While the whole message has not been written, alternate between
		// sendBLEPacket and recvBLEPacket, to simulate a real interaction.
		for counter = 0; completewr == false; counter++ {
			requestBuffer := bytes.NewBuffer(make([]byte, 0, mtu))
			rw := ble.NewResponseWriter(requestBuffer)
			cwr, err := sendBLEPacket(&off, msg, mtu, rw)
			if err != nil {
				t.Fatal(err)
			}
			completewr = cwr
			completerd = recvBLEPacket(&out, &ml, ble.NewRequest(nil, requestBuffer.Bytes(), 0))
		}

		// Test read completeness
		if completerd == false {
			t.Error("The write operation has been completed, but the read operation has not")
		}

		// Test received message length against sended one
		if len(out) != len(msg) {
			t.Errorf("The sent message is of length %d, but the received message is of length %d.", len(msg), len(out))
		}

		// Test received message data against sended one
		if bytes.Compare(msg, out) != 0 {
			t.Error("The sent message data differs from the received one.")
		}
	}
}
