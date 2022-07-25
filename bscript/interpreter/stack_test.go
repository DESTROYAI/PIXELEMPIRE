// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
)

// tstCheckScriptError ensures the type of the two passed errors are of the
// same type (either both nil or both of type Error) and their error codes
// match when not nil.
func tstCheckScriptError(gotErr, wantErr error) error {
	// Ensure the error code is of the expected type and the error
	// code matches the value specified in the test instance.
	if reflect.TypeOf(gotErr) != reflect.TypeOf(wantErr) {
		return fmt.Errorf("wrong error - got %T (%[1]v), want %T", gotErr, wantErr) //nolint:errorlint // test code
	}
	if gotErr == nil {
		return nil
	}

	// Ensure the want error type is a script error.
	werr := &errs.Error{}
	if ok := errors.As(wantErr, werr); !ok {
		return fmt.Errorf("unexpected test error type %T", wantErr) //nolint:errorlint // test code
	}

	// Ensure the err