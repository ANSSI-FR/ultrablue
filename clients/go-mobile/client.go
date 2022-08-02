// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package gomobile

import (
	"crypto/rsa"
	"math/big"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/go-attestation/attest"
)

type CredentialBlob struct {
	Secret     []byte
	Cred       []byte
	CredSecret []byte
}

func createRSAPublicKey(publicBytes []byte, exponent int) rsa.PublicKey {
	public := big.NewInt(0).SetBytes(publicBytes)
	key := rsa.PublicKey{
		N: public,
		E: exponent,
	}
	return key
}

func MakeCredential(ekn []byte, eke int, encodedAP []byte) *CredentialBlob {
	var ap attest.AttestationParameters

	err := cbor.Unmarshal(encodedAP, &ap)
	if err != nil {
		return &CredentialBlob{nil, nil, nil}
	}
	ekPub := createRSAPublicKey(ekn, eke)
	activationParams := attest.ActivationParameters{
		TPMVersion: attest.TPMVersion20,
		EK:         &ekPub,
		AK:         ap,
	}
	s, ec, err := activationParams.Generate()
	if err != nil {
		return &CredentialBlob{nil, nil, nil}
	}
	return &CredentialBlob{s, ec.Credential, ec.Secret}
}

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
