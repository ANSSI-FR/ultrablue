// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// TCG TPM v2.0 Provisioning Guidance, section 7.8:
// https://trustedcomputinggroup.org/wp-content/uploads/TCG-TPM-v2.0-Provisioning-Guidance-Published-v1r1.pdf
const SRK_HANDLE tpmutil.Handle = 0x81000001

// TCG TPM v2.0 Provisioning Guidance, section 7.5.1:
// https://trustedcomputinggroup.org/wp-content/uploads/TCG-TPM-v2.0-Provisioning-Guidance-Published-v1r1.pdf
// TCG EK Credential Profile, section 2.1.5.1:
// https://trustedcomputinggroup.org/wp-content/uploads/Credential_Profile_EK_V2.0_R14_published.pdf
var SRK_TEMPLATE = tpm2.Public {
	Type: tpm2.AlgRSA,
	NameAlg: tpm2.AlgSHA256,
	Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagDecrypt | tpm2.FlagNoDA,
	AuthPolicy: nil,
	RSAParameters: &tpm2.RSAParams {
		Symmetric: &tpm2.SymScheme {
			Alg: tpm2.AlgAES,
			KeyBits: 128,
			Mode: tpm2.AlgCFB,
		},
		KeyBits: 2048,
		ModulusRaw: make([]byte, 256),
	},
}

/*
	Returns a handle to the TPM Storage Rook Key (SRK) if it's already present
	in its non volatile memory.
	Otherwise, a new SRK is created using the default template, and is persisted
	in the TPM Non Volatile memory, at its default location. A handle to the newly
	created key is then returned
*/
func TPM2_LoadSRK(rwc io.ReadWriteCloser) (tpmutil.Handle, error) {
	// TPM2 Library commands, section 30.2:
	// https://trustedcomputinggroup.org/wp-content/uploads/TCG_TPM2_r1p59_Part3_Commands_pub.pdf
	// "TPM_CAP_HANDLES – Returns a list of all of the handles within the handle range
	// of the property parameter. The range of the returned handles is determined
	// by the handle type (the most-significant octet (MSO) of the property)."
	const PROPERTY = uint32(tpm2.HandleTypePersistent) << 24
	const MAX_OBJECTS = 256

	handles, _, err := tpm2.GetCapability(rwc, tpm2.CapabilityHandles, MAX_OBJECTS, PROPERTY)
	if err != nil {
		return 0, nil
	}
	for _, h := range handles {
		if h.(tpmutil.Handle) == SRK_HANDLE {
			return SRK_HANDLE, nil
		}
	}
	logrus.Info("SRK not found, creating one. This may take a while.")

	handle, _, err := tpm2.CreatePrimary(rwc, tpm2.HandleOwner, tpm2.PCRSelection{}, "", "", SRK_TEMPLATE)
	if err != nil {
		return 0, err
	}
	if err = tpm2.EvictControl(rwc, "", tpm2.HandleOwner, handle, SRK_HANDLE); err != nil {
		return 0, err
	}
	logrus.Infof("Persistent SRK created at NV index %x\n", SRK_HANDLE)
	return SRK_HANDLE, nil
}

/*
	Seals the given data to the TPM Storage Root Key (SRK) and to the
	provided PIN if the --with-pin command line argument is provided.
	Returns the resulting private and public blobs

	If the --with-pin command line argument is provided, a password policy
	will be used to seal the key.
	To get it back, the same policy will be needed at unseal time.
*/
func TPM2_Seal(data []byte, pin string) ([]byte, []byte, error) {
	var rwc io.ReadWriteCloser
	var priv, pub, policy []byte
	var srkHandle, sessHandle tpmutil.Handle
	var err error

	if rwc, err = tpm2.OpenTPM(); err != nil {
		return nil, nil, err
	}
	defer rwc.Close()

	if srkHandle, err = TPM2_LoadSRK(rwc); err != nil {
		return nil, nil, err
	}
	if sessHandle, _, err = tpm2.StartAuthSession(rwc, tpm2.HandleNull, tpm2.HandleNull, make([]byte, 16), nil, tpm2.SessionPolicy, tpm2.AlgNull, tpm2.AlgSHA256); err != nil {
		return nil, nil, err
	}

	// Note that we check for the command line argument rather than for an
	// empty PIN, because we don't want to transparently disable the password
	// policy for the session if the user inputs an empty string while providing
	// the --with-pin option.
	if *withpin {
		if err = tpm2.PolicyPassword(rwc, sessHandle); err != nil {
			return nil, nil, err
		}
	}

	if policy, err = tpm2.PolicyGetDigest(rwc, sessHandle); err != nil {
		return nil, nil, err
	}
	if priv, pub, err = tpm2.Seal(rwc, srkHandle, "", pin, policy, data); err != nil {
		return nil, nil, err
	}
	return priv, pub, err
}

/*
	Unseals the given object with the TPM Storage Root Key (SRK)
	and returns the original data, assuming the policy is correct:

	If the --with-pin command line argument is provided, a password policy is
	used for the session. This means that an incorrect PIN will increment
	the TPM DA counter, and may lock the TPM. This is wanted and prevents
	brute-force attacks.
	If the --with-pin was used at enroll time (thus sealing the key with a password
	policy), and it is not at attestation time: The password policy will not be
	used and TPM2_Unseal will return a policy error, without trying any password.
	In consequence, the TPM DA counter will NOT be incremented, which is also
	wanted because it is likely to be a usage error.
*/
func TPM2_Unseal(priv, pub []byte, pin string) ([]byte, error) {
	var rwc io.ReadWriteCloser
	var data []byte
	var srkHandle, keyHandle, sessHandle tpmutil.Handle
	var err error

	if rwc, err = tpm2.OpenTPM(); err != nil {
		return nil, err
	}
	defer rwc.Close()

	if srkHandle, err = TPM2_LoadSRK(rwc); err != nil {
		return nil, err
	}
	if sessHandle, _, err = tpm2.StartAuthSession(rwc, tpm2.HandleNull, tpm2.HandleNull, make([]byte, 16), nil, tpm2.SessionPolicy, tpm2.AlgNull, tpm2.AlgSHA256); err != nil {
		return nil, err
	}
	if keyHandle, _, err = tpm2.Load(rwc, srkHandle, "", pub, priv); err != nil {
		return nil, err
	}
	if *withpin {
		if err = tpm2.PolicyPassword(rwc, sessHandle); err != nil {
			return nil, err
		}
	}
	if data, err = tpm2.UnsealWithSession(rwc, sessHandle, keyHandle, pin); err != nil {
		return nil, err
	}
	return data, nil
}

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
	if rbytes == nil || len(rbytes) != int(size) || onlyContainsZeros(rbytes) {
		return nil, errors.New("Failed to generate random bytes")
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
