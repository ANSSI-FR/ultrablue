// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

/*
   Returns @size bytes of random data from the TPM
*/
func TPM2_GetRandom(size uint16) ([]byte, error) {
	rwc, err := tpm2.OpenTPM()
	if err != nil {
		return nil, err
	}
	defer rwc.Close()

	rbytes, err := tpm2.GetRandom(rwc, size)
	if err != nil {
		return nil, err
	}
	return rbytes, nil
}

/*
	Extends the PCR at the given index with @secret
	TODO: The following extends the PCR, but does not adds
	any event log entry. This is sufficient for our needs,
	but it means that after an ultrablue attestation, the
	eventlog will differ from the pcrs, thus making any
	replay fail.
	NOTE: It seems that it's not possible to add an event log
	entry directly from the tss stack, and that we need to use
	exposed UEFI function pointers.
*/
func TPM2_PCRExtend(index int, secret []byte) error {
	rwc, err := tpm2.OpenTPM()
	if err != nil {
		return err
	}
	defer rwc.Close()

	return tpm2.PCREvent(rwc, tpmutil.Handle(index), secret)
}
