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
	"strings"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/go-attestation/attest"
	"github.com/sergi/go-diff/diffmatchpatch"
)

/*
	CredentialBlob is a structure used during the
	credential activation process. It packs all
	return values of the ActivationParameters.generate
	method from the attest package.
*/
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

type EventLog struct {
	Raw []byte
}

type Diff struct {
	Raw []byte
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
func CheckQuotesSignature(encodedap, encodedpp, nonce []byte) error {
	var ap attest.AttestationParameters
	var pp attest.PlatformParameters

	// Decoding CBOR parameters
	if err := cbor.Unmarshal(encodedap, &ap); err != nil {
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

func isPrintable(data []byte) bool {
	for _, b := range data {
		if (b < 32 || b > 127) && b != 0 {
			return false
		}
	}
	return true
}

func GetParsedEventLog(encodedpp []byte) (*EventLog, error) {
	var pp attest.PlatformParameters
	var els = ""

	if err := cbor.Unmarshal(encodedpp, &pp); err != nil {
		return nil, err
	}
	el, err := attest.ParseEventLog(pp.EventLog)
	if err != nil {
		return nil, err
	}
	es := el.Events(attest.HashSHA256)
	for i, e := range es {
		els += fmt.Sprintln("- EventNum:", i)
		els += fmt.Sprintln("\tPCRIndex:", e.Index)
		els += fmt.Sprintln("\tEventType:", e.Type)
		els += fmt.Sprintln("\tDigest:", hex.EncodeToString(e.Digest))
		if isPrintable(e.Data) {
			els += fmt.Sprintln("\tData:", string(e.Data))
		}
	}
	return &EventLog { Raw: []byte(els) }, nil
}

func getLastLines(s string, count int) string {
	s = strings.Trim(s, "\n")
	lines := strings.Split(s, "\n")
	if count < len(lines) {
		lines = lines[len(lines) - count:]
	}
	return strings.Join(lines, "\n")
}

func getFirstLines(s string, count int) string {
	s = strings.Trim(s, "\n")
	lines := strings.Split(s, "\n")
	if count < len(lines) {
		lines = lines[:count]
	}
	return strings.Join(lines, "\n")
}

func prettyDiff(d diffmatchpatch.Diff) string {
	var prefix = map[diffmatchpatch.Operation]string {
		diffmatchpatch.DiffDelete: "<",
		diffmatchpatch.DiffEqual: " ",
		diffmatchpatch.DiffInsert: ">",
	}
	var diff = ""

	strs := strings.Split(strings.Trim(d.Text, "\n"), "\n")
	for _, s := range strs {
		diff += fmt.Sprintln(string(prefix[d.Type]), s)
	}
	return diff
}

func prettyDiffs(diffs []diffmatchpatch.Diff, contextLines int) string {
	var currentline int
	var diff = ""	

	var i int
	for i < len(diffs) {
		d := diffs[i]
		if d.Type != diffmatchpatch.DiffEqual {
			var si, ei, sl, el int // start index, end index, start line, end line
			si = i
			sl = currentline + 1
			for diffs[i].Type != diffmatchpatch.DiffEqual {
				currentline += strings.Count(diffs[i].Text, "\n")
				i += 1
			}
			ei = i
			el = currentline + 1
			diff += fmt.Sprintf("@@ l. %d:%d @@\n", sl, el)
			if si > 0 {
				diff += fmt.Sprintln(getLastLines(diffs[si - 1].Text, contextLines))
			}
			for j := si; j < ei; j++ {
				diff += prettyDiff(diffs[j])
			}
			if ei < len(diffs) {
				diff += fmt.Sprintln(getFirstLines(diffs[ei].Text, contextLines))
			}
			continue
		}
		currentline += strings.Count(d.Text, "\n")
		i += 1
	}
	return diff
}

func GetDiff(s1, s2 string) *Diff {
	dmp := diffmatchpatch.New()
	wSrc, wDst, warray := dmp.DiffLinesToRunes(s1, s2)
	diffs := dmp.DiffMainRunes(wSrc, wDst, false)
	diffs = dmp.DiffCharsToLines(diffs, warray)
	s := prettyDiffs(diffs, 3)
	return &Diff { Raw: []byte(s) }
}
