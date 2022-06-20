// SPDX-FileCopyrightText: 2022 ANSSI
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
)

/*
	usage prints informations to guide the user at the
	start of the program
*/
func usage(erl bool) {
	if erl {
		fmt.Println(colorYellow + `To enroll a new verifier device:

   1. Install the Ultrablue application on your smartphone
   2. From it, push the + button on the top-right corner
   3. Scan the following QR code
` + colorReset)
	} else {
		fmt.Println(colorYellow + `To perform the attestation:

   1. Open the Ultrablue application on an enrolled device
   2. Find this computer in the list of known attesters
   3. Tap the play button
` + colorReset)
	}
}

/*
	log displays a formatted message if the global @loglevel is >= than @level
*/
func log(level int, log ...interface{}) {
	if *loglevel >= level {
		switch level {
		case 0:
			fmt.Print("\n" + colorBlue + "[")
			fmt.Print(log...)
			fmt.Println("]" + colorReset + "\n")
		case 1:
			fmt.Print(colorGreen + "* ")
			fmt.Print(log...)
			fmt.Println(colorReset)
		default:
			fmt.Print("   - ")
			fmt.Println(log...)
		}
	}
}

/*
	Logs an error in red, with a "Fatal error: " prefix
*/
func logErr(err error) {
	fmt.Println(colorRed+"Fatal error:", err.Error()+colorReset)
}
