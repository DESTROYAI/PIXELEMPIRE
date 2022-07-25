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
// match