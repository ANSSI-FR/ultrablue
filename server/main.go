// The ultrablue-server program starts the Ultrablue server side application.
package main

import (
	"flag"
)

// Command line arguments (global variables)
var (
	tpmPath  = flag.String("tpm", "/dev/tpm0", "Path of the TPM device to use")
	enroll   = flag.Bool("enroll", false, "Must be true for a first time attestation")
	loglevel = flag.Int("loglevel", 1, "Indicates the level of logging, between 0 (no logs) and 2 (verbose)")
)

func main() {
	flag.Parse()
}
