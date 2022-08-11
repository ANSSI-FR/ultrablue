// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

/*
	This file contains go functions that are meant to be compiled
	for a mobile architecture, and binded to its native language.
	This way, those functions are available while developing
	native applicatioins (IOS/Android).

	We embeed go in mobile applications because some libraries
	(mainly go-attestation) don't exist in higher level languages,
	and it would take too much time to reimplement them.
*/

package gomobile

import (
	"crypto/rsa"
	"math/big"
	"errors"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/go-attestation/attest"
)

type CredentialBlob struct {
	Secret     []byte
	Cred       []byte
	CredSecret []byte
}

/*
	PCRs are returned encoded to CBOR because
	gobind can't return complex types such as
	arrays.
*/
type EncodedPCRs struct {
	Data []byte
}

/*
	buildRSAPublicKey rebuilds a crypto.PublicKey from raw public
	bytes and exponent of an RSA key.
*/
func buildRSAPublicKey(publicBytes []byte, exponent int) rsa.PublicKey {
	public := big.NewInt(0).SetBytes(publicBytes)
	key := rsa.PublicKey{
		N: public,
		E: exponent,
	}
	return key
}

/*
	MakeCredential generates a challenge used to assert that
	the attestation key @encodedap has been generated on the
	same TPM that the endorsement key @ekn + @eke.
*/
func MakeCredential(ekn []byte, eke int, encodedap []byte) (*CredentialBlob, error) {
	var ap attest.AttestationParameters

	err := cbor.Unmarshal(encodedap, &ap)
	if err != nil {
		return nil, err
	}
	ekPub := buildRSAPublicKey(ekn, eke)
	activationParams := attest.ActivationParameters{
		TPMVersion: attest.TPMVersion20,
		EK:         &ekPub,
		AK:         ap,
	}
	s, ec, err := activationParams.Generate()
	if err != nil {
		return nil, err
	}
	return &CredentialBlob{s, ec.Credential, ec.Secret}, nil
}

/*
	CheckQuotesSignature verifies that all quotes coming from
	the attestation data in @encodedpp are signed by the
	attestation key @encodedak, and contains the anti replay
	nonce.

	It also asserts that the final PCRs values from the attestation
	data matches the quotes ones.
*/
func CheckQuotesSignature(encodedak, encodedpp, nonce []byte) error {
	var ap attest.AttestationParameters
	var pp attest.PlatformParameters

	// Decoding CBOR parameters
	if err := cbor.Unmarshal(encodedak, &ap); err != nil {
		return err
	}
	if err := cbor.Unmarshal(encodedpp, &pp); err != nil {
		return err
	}

	// Verify attestation quotes
	akpub, err := attest.ParseAKPublic(attest.TPMVersion20, ap.Public)
	if err != nil {
		return err
	}
	for _, quote := range pp.Quotes {
		if err := akpub.Verify(quote, pp.PCRs, nonce); err != nil {
			return err
		}
	}
	return nil
}

/*
	ReplayEventLog verifies that the eventlog from @encodedpp
	matches the final PCRs values (that were previously checked
	against the quotes ones).
*/
func ReplayEventLog(encodedpp []byte) error {
	var pp attest.PlatformParameters

	if err := cbor.Unmarshal(encodedpp, &pp); err != nil {
		return err
	}
	el, err := attest.ParseEventLog(pp.EventLog)
	if err != nil {
		return err
	}
	_, err = el.Verify(pp.PCRs)
	if rErr, isReplayErr := err.(attest.ReplayError); isReplayErr {
		return errors.New(rErr.Error())
	}
	return err
}

/*
	GetPCRs extracts, encodes and returns PCRs from the
	attestation data @encodedpp.
*/
func GetPCRs(encodedpp []byte) (*EncodedPCRs, error) {
	var pp attest.PlatformParameters

	if err := cbor.Unmarshal(encodedpp, &pp); err != nil {
		return nil, err
	}
	ep, err := cbor.Marshal(pp.PCRs)
	if err != nil {
		return nil, err
	}
	return &EncodedPCRs {ep}, nil
}
