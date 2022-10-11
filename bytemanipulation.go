package bt

import "encoding/binary"

// ReverseBytes reverses the bytes (little endian/big endian).
// This is used when computing merkle trees in Bitcoin, for example.
func ReverseBytes(a []byte) []byte {
	tmp := make([]byte, len(a))
	cop