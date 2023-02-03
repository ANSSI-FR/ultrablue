// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"os"
	"reflect"
	"time"

	"github.com/google/go-attestation/attest"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const PCR_EXTENSION_INDEX = 9


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

// EnrollData contains the TPM's endorsement RSA public key
// with an optional certificate.
// It is used to deconstruct complex crypto.Certificate go type
// in order to encode and send it.
// It also contains a boolean @PCRExtend that indicates the new verifier
// it must generate a new secret to send back on attestation success.
type EnrollData struct {
	EKCert    []byte // x509 key certificate (one byte set to 0 if none)
	EKPub     []byte // Raw public key bytes
	EKExp     int    // Public key exponent
	PCRExtend bool   // Whether or not PCR_EXTENSION_INDEX must be extended on attestation success
}

// As encoding raw byte arrays to CBOR is not handled very well by
// most libraries out there, we encapsulate those in a one-field
// structure.
type Bytestring struct {
	Bytes []byte
}

func parseAttestEK(ek *attest.EK) (EnrollData, error) {
	if reflect.TypeOf(ek.Public).String() != "*rsa.PublicKey" {
		return EnrollData{}, errors.New("Invalid key type:" + reflect.TypeOf(ek.Public).String())
	}
	var c []byte = make([]byte, 0)
	if ek.Certificate != nil {
		c = ek.Certificate.Raw
	}
	var n = ek.Public.(*rsa.PublicKey).N.Bytes()
	var e = ek.Public.(*rsa.PublicKey).E
	return EnrollData{c, n, e, *pcrextend}, nil
}

func establishEncryptedSession(ch chan []byte) (*Session, error) {
	var data Bytestring
	var key []byte
	var session = NewSession(ch)
	var err error

	logrus.Info("Getting client UUID")
	if err = recvMsg(&data, session); err != nil {
		return nil, err
	}
	if session.uuid, err = uuid.FromBytes(data.Bytes); err != nil {
		close(ch)
		return nil, err
	}

	if *enroll {
		key = enrollkey
		enrollkey = nil
		logrus.Info("Saving UUID & encryption key")
		err = storeKey(session.uuid.String(), key)
	} else {
		logrus.Info("Fetching encryption key")
		key, err = loadKey(session.uuid.String())
	}
	if err != nil {
		close(ch)
		return nil, err
	}
	if err := session.StartEncryption(key); err != nil {
		close(ch)
		return nil, err
	}
	return session, nil
}

func enrollment(session *Session, tpm *attest.TPM) error {
	logrus.Info("Retrieving EK pub and EK cert")
	eks, err := tpm.EKs()
	if err != nil {
		close(session.ch)
		return err
	}
	logrus.Info("Sending enrollment data")

	// Any key should do the job in principle, use the first one
	// as we expect it to include a certificate.
	ek, err := parseAttestEK(&eks[0])
	if err != nil {
		close(session.ch)
		return nil
	}
	err = sendMsg(ek, session)
	if err != nil {
		return err
	}
	return nil
}

func authentication(session *Session) error {
	logrus.Info("Starting authentication process")
	logrus.Info("Generating nonce")
	rbytes, err := TPM2_GetRandom(16)
	if err != nil {
		close(session.ch)
		return err
	}
	nonce := Bytestring{rbytes}
	logrus.Info("Sending nonce")
	err = sendMsg(nonce, session)
	if err != nil {
		return err
	}
	logrus.Info("Getting nonce back")
	var rcvd_nonce Bytestring
	err = recvMsg(&rcvd_nonce, session)
	if err != nil {
		return err
	}
	logrus.Info("Verifying nonce")
	if bytes.Equal(nonce.Bytes, rcvd_nonce.Bytes) == false {
		close(session.ch)
		return errors.New("Authentication failure: nonces differ")
	}
	logrus.Info("The client is now authenticated")
	return nil
}

func credentialActivation(session *Session, tpm *attest.TPM) (*attest.AK, error) {
	logrus.Info("Generating AK")
	ak, err := tpm.NewAK(nil)
	if err != nil {
		close(session.ch)
		return nil, err
	}
	err = sendMsg(ak.AttestationParameters(), session)
	if err != nil {
		return nil, err
	}
	logrus.Info("Getting credential blob")
	var ec attest.EncryptedCredential
	err = recvMsg(&ec, session)
	if err != nil {
		return nil, err
	}
	logrus.Info("Decrypting credential blob")
	decrypted, err := ak.ActivateCredential(tpm, ec)
	if err != nil {
		close(session.ch)
		return nil, err
	}
	logrus.Info("Sending back decrypted credential blob")
	err = sendMsg(Bytestring{decrypted}, session)
	if err != nil {
		return nil, err
	}
	return ak, nil
}

func attestation(session *Session, tpm *attest.TPM, ak *attest.AK) error {
	logrus.Info("Getting anti replay nonce")
	var nonce Bytestring
	err := recvMsg(&nonce, session)
	if err != nil {
		return err
	}
	logrus.Info("Retrieving attestation plateform data")
	ap, err := tpm.AttestPlatform(ak, nonce.Bytes, nil)
	if err != nil {
		close(session.ch)
		return err
	}
	err = sendMsg(ap, session)
	if err != nil {
		return err
	}
	return nil
}

func response(session *Session) error {
	logrus.Info("Getting attestation response")
	var response struct  {
		Err        bool
		Secret     []byte
	}
	err := recvMsg(&response, session)
	if err != nil {
		return err
	}
	if response.Err {
		close(session.ch)
		return errors.New("Attestation failure")
	}
	if *enroll {
		logrus.Info("Enrollment success")
	} else {
		logrus.Info("Attestation success")
	}
	if len(response.Secret) > 0 {
		logrus.Info("Extending PCR", PCR_EXTENSION_INDEX)
		if err = TPM2_PCRExtend(PCR_EXTENSION_INDEX, response.Secret); err != nil {
			return err
		}
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
	var session *Session

	tpm, err := attest.OpenTPM(nil)
	if err != nil {
		logrus.Fatal(err)
	}
	defer tpm.Close()

	if session, err = establishEncryptedSession(ch); err != nil {
		logrus.Error(err); return
	}
	err = authentication(session)
	if err != nil {
		logrus.Error(err)
		return
	}
	if *enroll {
		err = enrollment(session, tpm)
		if err != nil {
			logrus.Error(err)
			return
		}
	}
	ak, err := credentialActivation(session, tpm)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = attestation(session, tpm, ak)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = response(session)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	time.Sleep(time.Second)
	os.Exit(0)
}
