// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/go-attestation/attest"
	"github.com/sirupsen/logrus"
)

const REGISTRATION_OK = 2

/*
	sendMsg takes the data to send, which is a generic,
	and sends it to the message channel of the connection
	state, encoded to CBOR.
	This will make the message available to read
	on the characteristic, and the function will
	block until the client reads it completely.
*/
func sendMsg[T any](data T, ch chan []byte) error {
	logrus.Debug("Encoding to CBOR")
	encoded, err := cbor.Marshal(data)
	if err != nil {
		close(ch)
		return err
	}
	logrus.Debug("Sending message")
	ch <- encoded
	_, ok := <-ch
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
*/
func recvMsg[T any](obj *T, ch chan []byte) error {
	logrus.Debug("Receiving message")
	rsp, ok := <-ch
	if !ok {
		return errors.New("The channel has been closed")
	}
	logrus.Debug("Decoding from CBOR")
	err := cbor.Unmarshal(rsp, obj)
	if err != nil {
		return err
	}
	return nil
}

/*
	ultrablueProtocol is the function that drives
	the server-client interaction, and implements the
	attestation protocol. It runs in a go routine, and
	closely cooperates with the ultrablueChr go-routine
	through the @ch channel. (As pointed out at the
	top of main.go, the BLE client has the control over
	the communication.)

	About error handling: When the error comes from the
	sendMsg/recvMsg methods, we can just return, and assume the
	connection has already been closed. When the
	error comes from the protocol, we need to first
	close the channel, to notify the characteristic that
	it needs to close the connection on the next client interaction.
*/
func ultrablueProtocol(ch chan []byte) {

	tpm, err := attest.OpenTPM(nil)
	if err != nil {
		logrus.Fatal(err)
	}

	if *enroll {
		logrus.Info("Retrieving EK pub and EK cert")
		eks, err := tpm.EKs()
		if err != nil {
			close(ch)
			return
		}

		err = sendMsg(&eks[0], ch)
		if err != nil {
			logrus.Error(err)
			return
		}

		logrus.Info("Getting registration confirmation")
		var rsp uint
		err = recvMsg(&rsp, ch)
		if err != nil {
			logrus.Error(err)
			return
		}
		if rsp != REGISTRATION_OK {
			return
		}
		logrus.Info("Registration success")
	}
}
