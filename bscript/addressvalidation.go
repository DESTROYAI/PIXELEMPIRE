package bscript

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/libsv/go-bk/crypto"
)

type a25 [25]byte

func (a *a25) embeddedChecksum() (c [4]byte) {
	copy(c[:], a[21:])
	return
}

// computeChecksum returns a four byte checksum computed from the first 21
// bytes of the address.  The embedded checksum is not updated.
func (a *a25) computeChecksum() (c [4]byte) {
	copy(c[:], crypto.Sha256d(a[:21]))
	return
}

// Tmpl and set58 are adapted from the 