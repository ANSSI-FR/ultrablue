// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/go-ble/ble"

var (
	ultrablueSvcUUID = ble.MustParse("ebee1789-50b3-4943-8396-16c0b7231cad")
	ultrablueChrUUID = ble.MustParse("ebee1790-50b3-4943-8396-16c0b7231cad")
)
