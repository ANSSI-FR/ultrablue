// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/go-ble/ble"

var (
	ultrablueSvcUUID = ble.MustParse("ebee1789-50b3-4943-8396-16c0b7231cad")

	registrationChrUUID         = ble.MustParse("ebee1790-50b3-4943-8396-16c0b7231cad")
	authenticationChrUUID       = ble.MustParse("ebee1791-50b3-4943-8396-16c0b7231cad")
	credentialActivationChrUUID = ble.MustParse("ebee1792-50b3-4943-8396-16c0b7231cad")
	attestationChrUUID          = ble.MustParse("ebee1793-50b3-4943-8396-16c0b7231cad")
	responseChrUUID             = ble.MustParse("ebee1794-50b3-4943-8396-16c0b7231cad")
)
