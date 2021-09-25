
package debug_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter"
	"github.com/libsv/go-bt/v2/bscript/interpreter/debug"
	"github.com/stretchr/testify/assert"
)

func TestDebugger_BeforeExecute(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStack           []string
		expOpcode          string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStack:           []string{},
			expOpcode:          "OP_4",
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStack:           []string{},
			expOpcode:          "OP_0",
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStack:           []string{},
			expOpcode:          "OP_4",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			lscript, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uscript, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			var timesCalled int
			debugger := debug.NewDebugger()
			debugger.AttachBeforeExecute(func(state *interpreter.State) {
				timesCalled++
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				assert.Equal(t, test.expStack, stack)
				assert.Equal(t, test.expOpcode, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(lscript, uscript),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, 1, timesCalled)
		})
	}
}

func TestDebugger_BeforeStep(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{},
				{"04"},
				{"04", "06"},
				{"04", "06", "02"},
				{"04", "06", "02", "03"},
				{"04", "06", "06"},
				{"04"},
				{"04", "02"},
				{"04", "02", "02"},
				{"04", "04"},
			},
			expOpcodes: []string{
				"OP_4", "OP_6",
				"OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY",
				"OP_2", "OP_2", "OP_ADD", "OP_EQUAL",
			},
		},
		"complex script": {
			lockingScriptHex:   "76a97ca8a687",
			unlockingScriptHex: "00",
			expStackHistory: [][]string{
				{},
				{""},
				{"", ""},
				{"", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", ""},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
				{"b472a266d0bd89c13706a4132ccfb16f7c3b9fcb", "b472a266d0bd89c13706a4132ccfb16f7c3b9fcb"},
			},
			expOpcodes: []string{"OP_0", "OP_DUP", "OP_HASH160", "OP_SWAP", "OP_SHA256", "OP_RIPEMD160", "OP_EQUAL"},
		},
		"error script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5457",
			expStackHistory: [][]string{
				{},
				{"04"},
				{"04", "07"},
				{"04", "07", "02"},
				{"04", "07", "02", "03"},
				{"04", "07", "06"},
			},
			expOpcodes: []string{"OP_4", "OP_7", "OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			lscript, err := bscript.NewFromHexString(test.lockingScriptHex)
			assert.NoError(t, err)

			uscript, err := bscript.NewFromHexString(test.unlockingScriptHex)
			assert.NoError(t, err)

			history := &stateHistory{
				dstack:  make([][]string, 0),
				astack:  make([][]string, 0),
				opcodes: make([]string, 0),
			}

			debugger := debug.NewDebugger()
			debugger.AttachBeforeStep(func(state *interpreter.State) {
				stack := make([]string, len(state.DataStack))
				for i, d := range state.DataStack {
					stack[i] = hex.EncodeToString(d)
				}
				history.dstack = append(history.dstack, stack)
				history.opcodes = append(history.opcodes, state.Opcode().Name())
			})

			interpreter.NewEngine().Execute(
				interpreter.WithScripts(lscript, uscript),
				interpreter.WithAfterGenesis(),
				interpreter.WithDebugger(debugger),
			)

			assert.Equal(t, test.expStackHistory, history.dstack)
			assert.Equal(t, test.expOpcodes, history.opcodes)
		})
	}
}

func TestDebugger_AfterStep(t *testing.T) {
	t.Parallel()

	type stateHistory struct {
		dstack  [][]string
		astack  [][]string
		opcodes []string
	}

	tests := map[string]struct {
		lockingScriptHex   string
		unlockingScriptHex string
		expStackHistory    [][]string
		expOpcodes         []string
	}{
		"simple script": {
			lockingScriptHex:   "5253958852529387",
			unlockingScriptHex: "5456",
			expStackHistory: [][]string{
				{"04"},
				{"04", "06"},
				{"04", "06", "02"},
				{"04", "06", "02", "03"},
				{"04", "06", "06"},
				{"04"},
				{"04", "02"},
				{"04", "02", "02"},
				{"04", "04"},
				{"01"},
			},
			expOpcodes: []string{
				"OP_6",
				"OP_2", "OP_3", "OP_MUL", "OP_EQUALVERIFY",