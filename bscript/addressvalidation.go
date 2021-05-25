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

// Tmpl and set58 are adapted from the C solution.
// Go has big integers but this technique seems better.
var tmpl = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// set58 takes a base58 encoded address and decodes it into the receiver.
// Errors are returned if the argument is not valid base58 or if the decoded
// value does not fit in the 25 byte address.  The address is not otherwise
// checked for validity.
func (a *a25) set58(s []byte) error {
	for _, s1 := range s {
		c := bytes.IndexByte(tmpl, s1)
		if c < 0 {
			return ErrEncodingBadChar
		}
		for