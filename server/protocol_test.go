// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"reflect"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestSendMsg_NormalCase(t *testing.T) {
	var data = []int{4, 8, 15, 16, 23, 42}
	var session = &Session {
		ch: make(chan []byte),
	}

	// Emulate a successful client read on the characteristic
	go func(ch chan []byte, t *testing.T) {
		_, ok := <-ch
		if !ok {
			t.Fatal("The channel has been closed")
		}
		ch <- nil
	}(session.ch, t)

	// Check that no error occured
	err := sendMsg(data, session)
	if err != nil {
		t.Errorf("Failed to send data %v", data)
	}
}

func TestSendMsg_WithError(t *testing.T) {
	var data = []int{4, 8, 15, 16, 23, 42}
	var session = &Session {
		ch: make(chan []byte),
	}

	// Emulate a client read on the characteristic that
	// fails , in a goroutine.
	go func(ch chan []byte, t *testing.T) {
		_, ok := <-ch
		if !ok {
			t.Fatal("The channel has been closed")
		}
		close(ch) // This means an error occured
	}(session.ch, t)

	// Make sure an error occured
	err := sendMsg(data, session)
	if err == nil {
		t.Error("sendMsg succeeded whereas the channel closed unexpectedly")
	}
}

/*
	// I'm sad to admit it, but I can't manage to make the cbor.Marshal function fail...
	func TestSendMsg_EncodingFailure(t *testing.T) {
		var data = []int {4, 8, 15, 16, 23, 42}
		var ch = make(chan []byte)

		// Emulate a client read on the characteristic
		go func(ch chan []byte, t *testing.T) {
			_, ok := <- ch
			if !ok {
				t.Fatal("The channel has been closed")
			}
			ch <- nil
		}(ch , t)

		// Make sure an error occured
		err := sendMsg(data, ch)
		if err == nil {
			t.Error("sendMsg succeeded whereas the channel closed unexpectedly")
		}
	}
*/

func TestRecvMsg_NormalCase(t *testing.T) {
	var data []int
	var expected = []int{4, 8, 15, 16, 23, 42}
	var session = &Session {
		ch: make(chan []byte),
	}

	// Emulate a successful client write on the characteristic
	go func(ch chan []byte, t *testing.T) {
		var data = []int{4, 8, 15, 16, 23, 42}
		var encoded, err = cbor.Marshal(data)
		if err != nil {
			t.Fatalf("Failed to encode %#v as CBOR", data)
		}
		ch <- encoded
	}(session.ch, t)

	// Check that no error occured
	err := recvMsg(&data, session)
	if err != nil {
		t.Error("Failed to receive data")
	}

	// Check the received data
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected: %#v, Received: %#v", expected, data)
	}
}

func TestRecvMsg_ChannelError(t *testing.T) {
	var data []int
	var session = &Session {
		ch: make(chan []byte),
	}

	// Emulate a channel error while a client is writing on
	// the characteristic.
	go func(ch chan []byte, t *testing.T) {
		close(ch)
	}(session.ch, t)

	// Make sure the error is caught
	err := recvMsg(&data, session)
	if err == nil {
		t.Error("recvMsg succeeded whereas the channel closed unecpectedly.")
	}
}

func TestRecvMsg_InvalidCBOR(t *testing.T) {
	var data []int
	var session = &Session {
		ch: make(chan []byte),
	}

	// Emulate a successful client write on the characteristic
	go func(ch chan []byte, t *testing.T) {
		var encoded = []byte{0x38, 0x18, 0x12} // Invalid CBOR
		ch <- encoded
	}(session.ch, t)

	// Make sure an error occured
	err := recvMsg(&data, session)
	if err == nil {
		t.Error("recvMsg succeeded whereas invalid CBOR data was sent")
	}
}
