// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
			t.Error("The channel has been closed")
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
			t.Error("The channel has been closed")
		}
		close(ch) // This means an error occured
	}(session.ch, t)

	// Make sure an error occured
	err := sendMsg(data, session)
	if err == nil {
		t.Error("sendMsg succeeded whereas the channel closed unexpectedly")
	}
}

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
			t.Errorf("Failed to encode %#v as CBOR", data)
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

func FakeCharacteristicRecv[T any](obj *T, session *Session) error {
	err := recvMsg(obj, session)
	if err != nil {
		return err
	}
	session.ch <- nil
	return nil
}

func FakeCharacteristicSend[T any](obj T, session *Session) error {
	err := sendMsg(obj, session)
	if err != nil {
		return err
	}
	return nil
}

func FakeCharacteristicRecvRaw(data []byte, session *Session) error {
	data, ok := <-session.ch
	if !ok {
		return errors.New("The channel has been closed")
	}
	session.ch <- nil
	return nil
}

func FakeCharacteristicSendRaw(data []byte, session *Session) error {
	session.ch <- data
	_, ok := <-session.ch
	if !ok {
		return errors.New("The channel has been closed")
	}
	return nil
}

func TestAuthentication_ValidNonce(t *testing.T) {
	logrus.SetLevel(0)
	var key = make([]byte, 32)
	var session = &Session {
		ch: make(chan []byte),
	}
	_, err := rand.Read(key)
	if err != nil {
		t.Error("Failed to generate random")
	}
	session.StartEncryption(key)

	// Emulate a genuine client
	go func(s *Session, t *testing.T) {
		var auth_nonce Bytestring
		err := FakeCharacteristicRecv(&auth_nonce, session)
		if err != nil {
			t.Error("Error while receiving auth nonce")
		}
		tweaked_nonce := Bytestring {
			Bytes: append(auth_nonce.Bytes[8:], auth_nonce.Bytes[:8]...),
		}
		if err = FakeCharacteristicSend(tweaked_nonce, session); err != nil {
			t.Error("Error while sending tweaked nonce")
		}
	}(session, t)

	if err := authentication(session); err != nil {
		t.Error("Unexpected error during authentication:", err)
	}
}

func TestAuthentication_UntweakedNonce(t *testing.T) {
	logrus.SetLevel(0)
	var key = make([]byte, 32)
	var session = &Session {
		ch: make(chan []byte),
	}
	_, err := rand.Read(key)
	if err != nil {
		t.Error("Failed to generate random")
	}
	session.StartEncryption(key)

	// Emulate a client that doesn't tweaks the nonce
	go func(s *Session, t *testing.T) {
		var auth_nonce Bytestring
		err := FakeCharacteristicRecv(&auth_nonce, session)
		if err != nil {
			t.Error("Error while receiving auth nonce")
		}
		auth_nonce = Bytestring {
			Bytes: make([]byte, 16),
		}
		FakeCharacteristicSend(auth_nonce, session)
	}(session, t)

	if err = authentication(session); err == nil {
		t.Error("An error should have occured, because the teak hasn't been applied", err)
	}
}

func TestAuthentication_ReplayedNonce(t *testing.T) {
	logrus.SetLevel(0)
	var session = &Session {
		ch: make(chan []byte),
	}

	// Emulate a client who doesn't have the symmetric key and attempts to resend
	// the unmodified reveived nonce
	go func(s *Session, t *testing.T) {
		var auth_nonce Bytestring
		err := FakeCharacteristicRecvRaw(auth_nonce.Bytes, session)
		if err != nil {
			t.Error("Error while receiving auth nonce")
		}
		FakeCharacteristicSendRaw(auth_nonce.Bytes, session)
	}(session, t)

	if err := authentication(session); err == nil {
		t.Error("An error should have occured, because the teak hasn't been applied", err)
	}
}
