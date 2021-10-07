
// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Package errs comment
package errs

import (
	"errors"
	"fmt"
)

// ErrorCode identifies a kind of script error.
type ErrorCode int

// These constants are used to identify a specific Error.
const (
	// ErrInternal is returned if internal consistency checks fail.  In
	// practice this error should never be seen as it would mean there is an
	// error in the engine logic.
	ErrInternal ErrorCode = iota

	// ErrOK represents successful execution. It should be treated similar to that
	// ok io.EOF
	ErrOK

	// ---------------------------------------
	// Failures related to improper API usage.
	// ---------------------------------------

	// ErrInvalidFlags is returned when the passed flags to NewEngine
	// contain an invalid combination.
	ErrInvalidFlags

	// ErrInvalidIndex is returned when an out-of-bounds index is passed to
	// a function.
	ErrInvalidIndex

	// ErrUnsupportedAddress is returned when a concrete type that
	// implements a bsvutil.Address is not a supported type.
	ErrUnsupportedAddress

	// ErrNotMultisigScript is returned from CalcMultiSigStats when the
	// provided script is not a multisig script.
	ErrNotMultisigScript

	// ErrTooManyRequiredSigs is returned from MultiSigScript when the
	// specified number of required signatures is larger than the number of
	// provided public keys.
	ErrTooManyRequiredSigs

	// ErrTooMuchNullData is returned from NullDataScript when the length of
	// the provided data exceeds MaxDataCarrierSize.
	ErrTooMuchNullData

	// ErrInvalidParams is returned when the ExectionParams passed to the Execute
	// func have errors.
	ErrInvalidParams

	// ------------------------------------------
	// Failures related to final execution state.
	// ------------------------------------------

	// ErrEarlyReturn is returned when OP_RETURN is executed in the script.
	ErrEarlyReturn

	// ErrEmptyStack is returned when the script evaluated without error,
	// but terminated with an empty top stack element.
	ErrEmptyStack

	// ErrEvalFalse is returned when the script evaluated without error but
	// terminated with a false top stack element.
	ErrEvalFalse

	// ErrScriptUnfinished is returned when CheckErrorCondition is called on
	// a script that has not finished executing.
	ErrScriptUnfinished

	// ErrInvalidProgramCounter is returned when an attempt to execute an opcode is
	// made once all of them have already been executed.  This can happen
	// due to things such as a second call to Execute or calling Step after
	// all opcodes have already been executed.
	ErrInvalidProgramCounter

	// -----------------------------------------------------
	// Failures related to exceeding maximum allowed limits.
	// -----------------------------------------------------

	// ErrScriptTooBig is returned if a script is larger than MaxScriptSize.
	ErrScriptTooBig

	// ErrElementTooBig is returned if the size of an element to be pushed
	// to the stack is over MaxScriptElementSize.
	ErrElementTooBig

	// ErrTooManyOperations is returned if a script has more than
	// MaxOpsPerScript opcodes that do not push data.
	ErrTooManyOperations

	// ErrStackOverflow is returned when stack and altstack combined depth
	// is over the limit.
	ErrStackOverflow

	// ErrInvalidPubKeyCount is returned when the number of public keys
	// specified for a multisig is either negative or greater than
	// MaxPubKeysPerMultiSig.
	ErrInvalidPubKeyCount

	// ErrInvalidSignatureCount is returned when the number of signatures
	// specified for a multisig is either negative or greater than the
	// number of public keys.
	ErrInvalidSignatureCount

	// ErrNumberTooBig is returned when the argument for an opcode that
	// expects numeric input is larger than the expected maximum number of
	// bytes.  For the most part, opcodes that deal with stack manipulation
	// via offsets, arithmetic, numeric comparison, and boolean logic are
	// those that this applies to.  However, any opcode that expects numeric
	// input may fail with this code.
	ErrNumberTooBig

	// ErrNumberTooSmall is returned when the argument for an opcode that