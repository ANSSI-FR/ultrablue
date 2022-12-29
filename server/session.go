// SPDX-FileCopyrightText: 2023 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

/*
	The Session type is an abstraction on top of the Go channel used to pass
	binary data to the goroutine reading and writing on the BLE characteristic.
	It takes care of automatically encoding and encrypting Go objects before
	sending them on the channel, and the other way around for received messages.
*/
type Session struct {
	ch chan []byte
	aesgcm cipher.AEAD
	encrypted bool
	uuid uuid.UUID
}

/*
	Creates and returns a new Session for the given channel
*/
func NewSession(ch chan []byte) *Session {
	return &Session {
		ch: ch,
	}
}

/*
	Creates an AES/GCM cipher from the given key and stores it
	in the Session. Also marks it as encrypted, so that
	subsequent calls to sendMsg/recvMsg with this Session will
	be encrypted.
*/
func (s *Session) StartEncryption(key []byte) error {
	if s.encrypted {
		return errors.New("The session is already encrypted")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	if s.aesgcm, err = cipher.NewGCM(block); err != nil {
		return err
	}
	s.encrypted = true
	return nil
}

/*
	sendMsg takes the data to send, which is a generic,
	and sends it to the message channel (through the session)
	of the connection state, encoded to CBOR.
	This will make the message available to read
	on the characteristic, and the function will
	block until the client reads it completely.

	If an error arises, and the channel is still open,
	sendMsg closes it.
*/
func sendMsg[T any](obj T, session *Session) error {
	logrus.Debug("Encoding to CBOR")
	data, err := cbor.Marshal(obj)
	if err != nil {
		close(session.ch)
		return err
	}
	if session.encrypted {
		logrus.Debug("Encrypting (AES/GCM)")
		iv, err := TPM2_GetRandom(uint16(session.aesgcm.NonceSize()))
		if err != nil {
			close(session.ch)
			return err
		}
		data = session.aesgcm.Seal(iv, iv, data, nil) // Append encrypted data to the IV
	}
	logrus.Debug("Sending message")
	session.ch <- data
	_, ok := <-session.ch
	if !ok {
		return errors.New("The channel has been closed")
	}
	return nil
}

/*
	recvMsg blocks until a message has been fully
	written by the client on the characteristic.
	It then tries to decode the CBOR message, and
	stores it in the obj parameter. Since obj is declared
	beforehand, and has a strong type, the cbor package
	will be able to decode it.

	If an error arises, and the channel is still open,
	recvMsg closes it.
*/
func recvMsg[T any](obj *T, session *Session) error {
	var err error

	logrus.Debug("Receiving message")
	data, ok := <-session.ch
	if !ok {
		return errors.New("The channel has been closed")
	}
	if session.encrypted {
		logrus.Debug("Decrypting (AES/GCM)")
		nonceSize := session.aesgcm.NonceSize()
		if len(data) < nonceSize {
			return errors.New("Data not large enough to be prefixed with the IV. Must be at least 12 bytes")
		}
		if data, err = session.aesgcm.Open(nil, data[:nonceSize], data[nonceSize:], nil); err != nil {
			close(session.ch)
			return err
		}
	}
	logrus.Debug("Decoding from CBOR")
	if err = cbor.Unmarshal(data, obj); err != nil {
		close(session.ch)
		return err
	}
	return nil
}
