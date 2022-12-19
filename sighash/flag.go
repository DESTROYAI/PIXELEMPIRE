// Package sighash comment
package sighash

// Flag represents hash type bits at the end of a signature.
type Flag uint8

// SIGHASH type bits from the end of a signature.
// see: https://wiki.bitcoinsv.io/index.php/SIGHASH_flags
const (
	Old          Flag = 0x0
	All          Flag = 0x1
	None         Flag = 0x2
	Single       Flag = 0x3
	AnyOneCanPay Flag = 0x80

	// Currently, all BitCoin (