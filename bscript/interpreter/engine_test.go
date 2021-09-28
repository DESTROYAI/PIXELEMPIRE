// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package interpreter

import (
	"errors"
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
	"github.com/libsv/go-bt/v2/sighash"
	"github.com/stretchr/testify/assert"
)

// TestBadPC sets the pc to a deliberately bad result then confirms that Step()
// and Disasm fail correctly.
func TestBadPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		script, off int
	}{
		{script: 2, off: 0},
		{script: 0, off: 2},
	}

	uscript, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to create unlocking script %e", err)
	}

	tx := &bt.Tx{
		Version: 1,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: 0,
			UnlockingScript:    uscript,
			SequenceNumber:     4294967295,
		}},
		Outputs: []*bt.Output{{
			Satoshis: 1000000000,
		}},
		LockTime: 0,
	}

	lscript, err := bscript.NewFromASM("OP_NOP")
	if err != nil {
		t.Errorf("failed to created locking script %e", err)
	}
	txOut := &bt.Output{
		LockingScript: lscript,
	}

	for _, test := range tests {
		vm := &thread{
			scriptParser: &DefaultOpcodeParser{},
			cfg:          &beforeGenesisConfig{},
		}
		err := vm.apply(&execOpts{
			previousTxOut: txOut,
			tx:            tx,
			inputIdx:      0,
		})
		if err != nil {
			t.Errorf("Failed to create script: %v", err)
		}

		// set to after all scripts
		vm.scriptIdx = test.script
		vm.scriptOff = test.off

		_, err = vm.Step()
		if err == nil {
			t.Errorf("Step with invalid pc (%v) succeeds!", test)
			continue
		}

		if err == nil {
			t.Errorf("DisasmPC with invalid pc (%v) succeeds!",
				test)
		}
	}
}

// TestCheckErrorCondition tests to execute early test in CheckErrorCondition()
// since most code paths are tested elsewhere.
func TestCheckErrorCondition(t *testing.T) {
	t.Parallel()

	tx := &bt.Tx{
		Version: 1,
		Inputs: []*bt.Input{{
			PreviousTxOutIndex: 0,
			UnlockingScript:    &bscript.Script{},
			SequenceNumber:     4294967295,
		}},
		Outputs: []*bt.Output{{
			Satoshis: 1000000000,
		}},
		LockTime: 0,
	}

	lscript, err := bscript.NewFromASM("OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_NOP OP_TRUE")
	if err != nil {
		t.Errorf("failed to created locking script %e", err)
	}
	txOut := &bt.Output{
		LockingScript: lscript,
	}

	vm := &thread{
		scriptParser: &DefaultOpcodeParser{},
		cfg:          &beforeGenesisConfig{},
	}

	err = vm.apply(&execOpts{
		previousTxOut: txOut,
		inputIdx:      0,
		tx:            tx,
	})
	if err != nil {
		t.Errorf("failed to configure thread %v", err)
	}

	var done bool
	for i := 0; i < len(*lscript); i++ {
		done, err = vm.Step()
		if err != nil {
			t.Fatalf("failed to step %dth time: %v", i, err)
		}
		if done && i != len(*lscript)-1 {
			t.Fatalf("finished early on %dth time", i)
		}
	}
	err = vm.CheckErrorCondition(false)
	if err != nil {
		t.Errorf("unexpected error %v on final check", err)
	}
}

func TestValidateParams(t *testing.T) {
	tests := map[string]struct {
		params execOpts
		expErr error
	}{
		"valid tx/previous out checksig script": {
			params: execOpts{
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac", 0)
					assert.NoError(t, err)

					uscript, err := bscript.NewFromHexString("483045022100a4d9da733aeb29f9ba94dcaa578e71662cf29dd9742ce4b022c098211f4fdb06022041d24db4eda239fa15a12cf91229f6c352adab3c1c10091fc2aa517fe0f487c5412102454c535854802e5eaeaf5cbecd20e0aa508486063b71194dfde34744f19f1a5d")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = uscript

					return tx
				}(),
				previousTxOut: func() *bt.Output {
					cbLockingScript, err := bscript.NewFromHexString("76a91454807ccc44c0eec0b0e187b3ce0e137e9c6cd65d88ac")
					assert.NoError(t, err)

					return &bt.Output{LockingScript: cbLockingScript, Satoshis: 0}
				}(),
			},
		},
		"valid tx/previous out non-checksig script": {
			params: execOpts{
				tx: func() *bt.Tx {
					tx := bt.NewTx()
					err := tx.From("ae81577c1a2434929a1224cf19aa63e167d88029965e2ca6de24defff014d031", 0, "52529387", 0)
					assert.NoError(t, err)

					txUnlockingScript, err := bscript.NewFromASM("OP_4")
					assert.NoError(t, err)

					tx.Inputs[0].UnlockingScript = txUnlockingScript

					return tx
				}(),
				previousTxOut: func() *bt.Output {
					cbLockingScript, 