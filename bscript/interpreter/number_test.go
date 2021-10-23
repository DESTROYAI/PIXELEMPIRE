
// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"bytes"
	"encoding/hex"
	"math"
	"math/big"
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
)

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error.  This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// TestScriptNumBytes ensures that converting from integral script numbers to
// byte representations works as expected.
func TestScriptNumBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		num        int64
		serialised []byte
	}{
		{0, nil},
		{1, hexToBytes("01")},
		{-1, hexToBytes("81")},
		{127, hexToBytes("7f")},
		{-127, hexToBytes("ff")},
		{128, hexToBytes("8000")},
		{-128, hexToBytes("8080")},
		{129, hexToBytes("8100")},
		{-129, hexToBytes("8180")},
		{256, hexToBytes("0001")},
		{-256, hexToBytes("0081")},
		{32767, hexToBytes("ff7f")},
		{-32767, hexToBytes("ffff")},
		{32768, hexToBytes("008000")},
		{-32768, hexToBytes("008080")},
		{65535, hexToBytes("ffff00")},
		{-65535, hexToBytes("ffff80")},
		{524288, hexToBytes("000008")},
		{-524288, hexToBytes("000088")},
		{7340032, hexToBytes("000070")},
		{-7340032, hexToBytes("0000f0")},
		{8388608, hexToBytes("00008000")},
		{-8388608, hexToBytes("00008080")},
		{2147483647, hexToBytes("ffffff7f")},
		{-2147483647, hexToBytes("ffffffff")},

		// Values that are out of range for data that is interpreted as
		// numbers, but are allowed as the result of numeric operations.
		{2147483648, hexToBytes("0000008000")},
		{-2147483648, hexToBytes("0000008080")},
		{2415919104, hexToBytes("0000009000")},
		{-2415919104, hexToBytes("0000009080")},
		{4294967295, hexToBytes("ffffffff00")},
		{-4294967295, hexToBytes("ffffffff80")},
		{4294967296, hexToBytes("0000000001")},
		{-4294967296, hexToBytes("0000000081")},
		{281474976710655, hexToBytes("ffffffffffff00")},
		{-281474976710655, hexToBytes("ffffffffffff80")},
		{72057594037927935, hexToBytes("ffffffffffffff00")},
		{-72057594037927935, hexToBytes("ffffffffffffff80")},
		{9223372036854775807, hexToBytes("ffffffffffffff7f")},
		{-9223372036854775807, hexToBytes("ffffffffffffffff")},
	}

	for _, test := range tests {
		n := &scriptNumber{val: big.NewInt(test.num)}
		if !bytes.Equal(n.Bytes(), test.serialised) {
			t.Errorf("Bytes: did not get expected bytes for %d - got %x, want %x", test.num, n.Bytes(), test.serialised)
			continue
		}
	}
}

// TestMakeScriptNum ensures that converting from byte representations to
// integral script numbers works as expected.
func TestMakeScriptNum(t *testing.T) {
	t.Parallel()

	// Errors used in the tests below defined here for convenience and to
	// keep the horizontal test size shorter.
	errNumTooBig := errs.NewError(errs.ErrNumberTooBig, "")
	errMinimalData := errs.NewError(errs.ErrMinimalData, "")

	tests := []struct {
		serialised      []byte
		num             int
		numLen          int
		minimalEncoding bool
		err             error
	}{
		// Minimal encoding must reject negative 0.
		{hexToBytes("80"), 0, MaxScriptNumberLengthBeforeGenesis, true, errMinimalData},

		// Minimally encoded valid values with minimal encoding flag.
		// Should not error and return expected integral number.
		{nil, 0, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("01"), 1, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("81"), -1, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("7f"), 127, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("ff"), -127, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("8000"), 128, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("8080"), -128, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("8100"), 129, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("8180"), -129, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("0001"), 256, MaxScriptNumberLengthBeforeGenesis, true, nil},
		{hexToBytes("0081"), -256, MaxScriptNumberLengthBeforeGenesis, true, nil},