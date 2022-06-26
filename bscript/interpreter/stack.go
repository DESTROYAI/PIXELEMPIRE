// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"encoding/hex"

	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
)

// asBool gets the boolean value of the byte array.
func asBool(t []byte) bool {
	for i := range t {
		if t[i] != 0 {
			// Negative 0 is also considered false.
			if i == len(t)-1 && t[i] == 0x80 {
				return false
			}
			return true
		}
	}
	return false
}

// fromBool converts a boolean into the appropriate byte array.
func fromBool(v bool) []byte {
	if v {
		return []byte{1}
	}
	return nil
}

// stack represents a stack of immutable objects to be used with bitcoin
// scripts.  Objects may be shared, therefore in usage if a value is to be
// changed it *must* be deep-copied first to avoid changing other values on the
// stack.
type stack struct {
	stk               [][]byte
	maxNumLength      int
	afterGenesis      bool
	verifyMinimalData bool
	debug             Debugger
	sh                StateHandler
}

func newStack(cfg config, verifyMinimalData bool) stack {
	return stack{
		maxNumLength:      cfg.MaxScriptNumberLength(),
		afterGenesis:      cfg.AfterGenesis(),
		verifyMinimalData: verifyMinimalData,
		debug:             &nopDebugger{},
		sh:                &nopStateHandler{},
	}
}

// Depth returns the number of items on the stack.
func (s *stack) Depth() int32 {
	return int32(len(s.stk))
}

// PushByteArray adds the given back array to the top of the stack.
//
// Stack transformation: [... x1 x2] -> [... x1 x2 data]
func (s *stack) PushByteArray(so []byte) {
	defer s.afterStackPush(so)
	s.beforeStackPush(so)
	s.stk = append(s.stk, so)
}

// PushInt converts the provided scriptNumber to a suitable byte array then pushes
// it onto the top of the stack.
//
// Stack transformation: [... x1 x2] -> [... x1 x2 int]
func (s *stack) PushInt(n *scriptNumber) {
	s.PushByteArray(n.Bytes())
}

// PushBool converts the provided boolean to a suitable byte array then pushes
// it onto the top of the stack.
//
// Stack transformation: [... x1 x2] -> [... x1 x2 bool]
func (s *stack) PushBool(val bool) {
	s.PushByteArray(fromBool(val))
}

// PopByteArray pops the value off the top of the stack and returns it.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2]
func (s *stack) PopByteArray() ([]byte, error) {
	s.beforeStackPop()
	data, err := s.nipN(0)
	if err != nil {
		return nil, err
	}
	s.afterStackPop(data)
	return data, nil
}

// PopInt pops the value off the top of the stack, converts it into a scriptNumber,
// and returns it.  The act of converting to a script num enforces the
// consensus rules imposed on data interpreted as numbers.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2]
func (s *stack) PopInt() (*scriptNumber, error) {
	so, err := s.PopByteArray()
	if err != nil {
		return nil, err
	}

	return makeScriptNumber(so, s.maxNumLength, s.verifyMinimalData, s.afterGenesis)
}

// PopBool pops the value off the top of the stack, converts it into a bool, and
// returns it.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2]
func (s *stack) PopBool() (bool, error) {
	so, err := s.PopByteArray()
	if err != nil {
		return false, err
	}

	return asBool(so), nil
}

// PeekByteArray returns the Nth item on the stack without removing it.
func (s *stack) PeekByteArray(idx int32) ([]byte, error) {
	sz := int32(len(s.stk))
	if idx < 0 || idx >= sz {
		return nil, errs.NewError(errs.ErrInvalidStackOperati