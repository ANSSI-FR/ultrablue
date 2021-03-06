// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/go-attestation/attest"
	"github.com/sirupsen/logrus"
)

/*
	sendMsg takes the data to send, which is a generic,
	and sends it to the message channel of the connection
	state, encoded to CBOR.
	This will make the message available to read
	on the characteristic, and the function will
	block until the client reads it completely.

	If an error arises, and the channel is still open,
	sendMsg closes it.
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

	If an error arises, and the channel is still open,
	recvMsg closes it.
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
		close(ch)
		return err
	}
	return nil
}

// ------------- PROTOCOL FUNCTIONS ---------------- //

/*
	Note on error handling: In the following functions, the channel
	is closed on some errors, but not on others.
	We must close the channel on error only if it is
	not already closed, that is, on errors that comes from
	the application layer.
	When the error comes from the transport layer (the
	sendMsg/recvMsg functions), the channel has already been
	closed, and the program will panic if we try to close it
	again.
*/

func registration(ch chan []byte, tpm *attest.TPM) error {
	logrus.Info("Retrieving EK pub and EK cert")
	eks, err := tpm.EKs()
	if err != nil {
		close(ch)
		return err
	}
	logrus.Info("Sending EK pub and EK cert")
	// Any key should do the job in principle, use the first one
	// as we expect it to include a certificate.
	err = sendMsg(&eks[0], ch)
	if err != nil {
		return err
	}
	logrus.Info("Getting registration confirmation")
	var regerr error
	err = recvMsg(&regerr, ch)
	if err != nil {
		return err
	}
	if regerr != nil {
		close(ch)
		return regerr
	}
	logrus.Info("Registration success")
	return nil
}

func authentication(ch chan []byte) error {
	logrus.Info("Starting authentication process")
	logrus.Info("Generating nonce")
	nonce, err := getTPMRandom(16)
	if err != nil {
		close(ch)
		return err
	}
	logrus.Info("Sending nonce")
	err = sendMsg(nonce, ch)
	if err != nil {
		return err
	}
	logrus.Info("Getting nonce back")
	var rcvd_nonce []byte
	err = recvMsg(&rcvd_nonce, ch)
	if err != nil {
		return err
	}
	logrus.Info("Verifying nonce")
	if bytes.Equal(nonce, rcvd_nonce) == false {
		close(ch)
		return errors.New("Authentication failure: nonces differ")
	}
	logrus.Info("The client is now authenticated")
	return nil
}

func credentialActivation(ch chan []byte, tpm *attest.TPM) (*attest.AK, error) {
	logrus.Info("Generating AK")
	ak, err := tpm.NewAK(nil)
	if err != nil {
		close(ch)
		return nil, err
	}
	err = sendMsg(ak.AttestationParameters(), ch)
	if err != nil {
		return nil, err
	}
	logrus.Info("Getting credential blob")
	var ec attest.EncryptedCredential
	err = recvMsg(&ec, ch)
	if err != nil {
		return nil, err
	}
	logrus.Info("Decrypting credential blob")
	decrypted, err := ak.ActivateCredential(tpm, ec)
	if err != nil {
		close(ch)
		return nil, err
	}
	logrus.Info("Sending back decrypted credential blob")
	err = sendMsg(decrypted, ch)
	if err != nil {
		return nil, err
	}
	return ak, nil
}

func attestation(ch chan []byte, tpm *attest.TPM, ak *attest.AK) error {
	logrus.Info("Getting anti replay nonce")
	var nonce []byte
	err := recvMsg(&nonce, ch)
	if err != nil {
		return err
	}
	logrus.Info("Retrieving attestation plateform data")
	ap, err := tpm.AttestPlatform(ak, nonce, nil)
	if err != nil {
		close(ch)
		return err
	}
	err = sendMsg(ap, ch)
	if err != nil {
		return err
	}
	var atterr error
	err = recvMsg(&atterr, ch)
	if err != nil {
		return err
	}
	if atterr != nil {
		close(ch)
		return atterr
	}
	logrus.Info("Attestation success")
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
	defer tpm.Close()
	if *enroll {
		err = registration(ch, tpm)
		if err != nil {
			logrus.Error(err)
			return
		}
	}
	err = authentication(ch)
	if err != nil {
		logrus.Error(err)
		return
	}
	ak, err := credentialActivation(ch, tpm)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = attestation(ch, tpm, ak)
	if err != nil {
		logrus.Error(err)
		return
	}
	os.Exit(0)
}
