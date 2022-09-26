
package interpreter

import (
	"math/big"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
	"github.com/libsv/go-bt/v2/sighash"
)

// halfOrder is used to tame ECDSA malleability (see BIP0062).
var halfOrder = new(big.Int).Rsh(bec.S256().N, 1)

type thread struct {
	dstack stack // data stack
	astack stack // alt stack

	elseStack boolStack

	cfg config

	debug Debugger
	state StateHandler

	scripts         []ParsedScript
	condStack       []int
	savedFirstStack [][]byte // stack from first script for bip16 scripts

	scriptParser OpcodeParser
	scriptIdx    int
	scriptOff    int
	lastCodeSep  int

	tx         *bt.Tx
	inputIdx   int
	prevOutput *bt.Output

	numOps int

	flags scriptflag.Flag
	bip16 bool // treat execution as pay-to-script-hash

	afterGenesis            bool
	earlyReturnAfterGenesis bool
}

func createThread(opts *execOpts) (*thread, error) {
	th := &thread{
		scriptParser: &DefaultOpcodeParser{
			ErrorOnCheckSig: opts.tx == nil || opts.previousTxOut == nil,
		},
		cfg: &beforeGenesisConfig{},
	}

	if err := th.apply(opts); err != nil {
		return nil, err
	}

	return th, nil
}

// execOpts are the params required for building an Engine
//
// Raw *bscript.Scripts can be supplied as LockingScript and UnlockingScript, or
// a Tx, an input index, and a previous output.
//
// If checksig operaitons are to be executed without a Tx or a PreviousTxOut supplied,
// the engine will return an ErrInvalidParams on execute.
type execOpts struct {
	lockingScript   *bscript.Script
	unlockingScript *bscript.Script
	previousTxOut   *bt.Output
	tx              *bt.Tx
	inputIdx        int
	flags           scriptflag.Flag
	debugger        Debugger
	state           *State
}

func (o execOpts) validate() error {
	// The provided transaction input index must refer to a valid input.
	if o.inputIdx < 0 || (o.tx != nil && o.inputIdx > o.tx.InputCount()-1) {
		return errs.NewError(
			errs.ErrInvalidIndex,
			"transaction input index %d is negative or >= %d", o.inputIdx, len(o.tx.Inputs),
		)
	}

	outputHasLockingScript := o.previousTxOut != nil && o.previousTxOut.LockingScript != nil
	txHasUnlockingScript := o.tx != nil && o.tx.Inputs != nil && len(o.tx.Inputs) > 0 &&
		o.tx.Inputs[o.inputIdx] != nil && o.tx.Inputs[o.inputIdx].UnlockingScript != nil
	// If no locking script was provided
	if o.lockingScript == nil && !outputHasLockingScript {
		return errs.NewError(errs.ErrInvalidParams, "no locking script provided")
	}

	// If no unlocking script was provided
	if o.unlockingScript == nil && !txHasUnlockingScript {
		return errs.NewError(errs.ErrInvalidParams, "no unlocking script provided")
	}

	// If both a locking script and previous output were provided, make sure the scripts match
	if o.lockingScript != nil && outputHasLockingScript {
		if !o.lockingScript.Equals(o.previousTxOut.LockingScript) {
			return errs.NewError(
				errs.ErrInvalidParams,
				"locking script does not match the previous outputs locking script",
			)
		}
	}

	// If both a unlocking script and an input were provided, make sure the scripts match
	if o.unlockingScript != nil && txHasUnlockingScript {
		if !o.unlockingScript.Equals(o.tx.Inputs[o.inputIdx].UnlockingScript) {
			return errs.NewError(
				errs.ErrInvalidParams,
				"unlocking script does not match the unlocking script of the requested input",
			)
		}
	}

	return nil
}

// hasFlag returns whether the script engine instance has the passed flag set.
func (t *thread) hasFlag(flag scriptflag.Flag) bool {
	return t.flags.HasFlag(flag)
}

func (t *thread) hasAny(ff ...scriptflag.Flag) bool {
	return t.flags.HasAny(ff...)
}

func (t *thread) addFlag(flag scriptflag.Flag) {
	t.flags.AddFlag(flag)
}

// isBranchExecuting returns whether the current conditional branch is
// actively executing. For example, when the data stack has an OP_FALSE on it
// and an OP_IF is encountered, the branch is inactive until an OP_ELSE or
// OP_ENDIF is encountered.  It properly handles nested conditionals.
func (t *thread) isBranchExecuting() bool {
	return len(t.condStack) == 0 || t.condStack[len(t.condStack)-1] == opCondTrue
}

// executeOpcode performs execution on the passed opcode. It takes into account
// whether it is hidden by conditionals, but some rules still must be
// tested in this case.
func (t *thread) executeOpcode(pop ParsedOpcode) error {
	if len(pop.Data) > t.cfg.MaxScriptElementSize() {
		return errs.NewError(errs.ErrElementTooBig,
			"element size %d exceeds max allowed size %d", len(pop.Data), t.cfg.MaxScriptElementSize())
	}

	exec := t.shouldExec(pop)

	// Disabled opcodes are fail on program counter.
	if pop.IsDisabled() && (!t.afterGenesis || exec) {
		return errs.NewError(errs.ErrDisabledOpcode, "attempt to execute disabled opcode %s", pop.Name())
	}

	// Always-illegal opcodes are fail on program counter.
	if pop.AlwaysIllegal() && !t.afterGenesis {
		return errs.NewError(errs.ErrReservedOpcode, "attempt to execute reserved opcode %s", pop.Name())
	}

	// Note that this includes OP_RESERVED which counts as a push operation.
	if pop.op.val > bscript.Op16 {
		t.numOps++
		if t.numOps > t.cfg.MaxOps() {
			return errs.NewError(errs.ErrTooManyOperations, "exceeded max operation limit of %d", t.cfg.MaxOps())
		}

	}

	if len(pop.Data) > t.cfg.MaxScriptElementSize() {
		return errs.NewError(errs.ErrElementTooBig,
			"element size %d exceeds max allowed size %d", len(pop.Data), t.cfg.MaxScriptElementSize())
	}

	// Nothing left to do when this is not a conditional opcode, and it is
	// not in an executing branch.
	if !t.isBranchExecuting() && !pop.IsConditional() {
		return nil
	}

	// Ensure all executed data push opcodes use the minimal encoding when
	// the minimal data verification flag is set.
	if t.dstack.verifyMinimalData && t.isBranchExecuting() && pop.op.val <= bscript.OpPUSHDATA4 && exec {
		if err := pop.enforceMinimumDataPush(); err != nil {
			return err
		}
	}

	// If we have already reached an OP_RETURN, we don't execute the next comment, unless it is a conditional,
	// in which case we need to evaluate it as to check for correct if/else balances
	if !exec && !pop.IsConditional() {
		return nil
	}

	return pop.op.exec(&pop, t)
}

// validPC returns an error if the current script position is valid for
// execution, nil otherwise.
func (t *thread) validPC() error {
	if t.scriptIdx >= len(t.scripts) {
		return errs.NewError(errs.ErrInvalidProgramCounter,
			"past input scripts %v:%v %v:xxxx", t.scriptIdx, t.scriptOff, len(t.scripts))
	}
	if t.scriptOff >= len(t.scripts[t.scriptIdx]) {
		return errs.NewError(errs.ErrInvalidProgramCounter, "past input scripts %v:%v %v:%04d", t.scriptIdx, t.scriptOff,
			t.scriptIdx, len(t.scripts[t.scriptIdx]))
	}
	return nil
}

// CheckErrorCondition returns nil if the running script has ended and was
// successful, leaving a true boolean on the stack.  An error otherwise,
// including if the script has not finished.
func (t *thread) CheckErrorCondition(finalScript bool) error {
	if t.dstack.Depth() < 1 {
		return errs.NewError(errs.ErrEmptyStack, "stack empty at end of script execution")
	}

	if finalScript && t.hasFlag(scriptflag.VerifyCleanStack) && t.dstack.Depth() != 1 {
		return errs.NewError(errs.ErrCleanStack, "stack contains %d unexpected items", t.dstack.Depth()-1)
	}

	v, err := t.dstack.PopBool()
	if err != nil {
		return err
	}
	if !v {
		return errs.NewError(errs.ErrEvalFalse, "false stack entry at end of script execution")
	}

	if finalScript {
		t.afterSuccess()
	}

	return nil
}

func (t *thread) apply(opts *execOpts) error {
	if err := opts.validate(); err != nil {
		return err
	}

	if opts.unlockingScript == nil {
		opts.unlockingScript = opts.tx.Inputs[opts.inputIdx].UnlockingScript
	}
	if opts.lockingScript == nil {
		opts.lockingScript = opts.previousTxOut.LockingScript
	}

	t.tx = opts.tx
	t.flags = opts.flags
	t.inputIdx = opts.inputIdx
	t.prevOutput = opts.previousTxOut

	// The clean stack flag (ScriptVerifyCleanStack) is not allowed without
	// the pay-to-script-hash (P2SH) evaluation (ScriptBip16).
	//
	// Recall that evaluating a P2SH script without the flag set results in
	// non-P2SH evaluation which leaves the P2SH inputs on the stack.
	// Thus, allowing the clean stack flag without the P2SH flag would make
	// it possible to have a situation where P2SH would not be a soft fork
	// when it should be.
	if t.hasFlag(scriptflag.EnableSighashForkID) {
		t.addFlag(scriptflag.VerifyStrictEncoding)
	}

	t.elseStack = &nopBoolStack{}
	if t.hasFlag(scriptflag.UTXOAfterGenesis) {
		t.elseStack = &stack{debug: &nopDebugger{}, sh: &nopStateHandler{}}
		t.afterGenesis = true
		t.cfg = &afterGenesisConfig{}
	}

	uscript := opts.unlockingScript
	lscript := opts.lockingScript

	// When both the signature script and public key script are empty the
	// result is necessarily an error since the stack would end up being
	// empty which is equivalent to a false top element.  Thus, just return
	// the relevant error now as an optimization.
	if (uscript == nil || len(*uscript) == 0) && (lscript == nil || len(*lscript) == 0) {
		return errs.NewError(errs.ErrEvalFalse, "false stack entry at end of script execution")
	}

	if t.hasFlag(scriptflag.VerifyCleanStack) && !t.hasFlag(scriptflag.Bip16) {
		return errs.NewError(errs.ErrInvalidFlags, "invalid scriptflag combination")
	}

	if len(*uscript) > t.cfg.MaxScriptSize() {
		return errs.NewError(
			errs.ErrScriptTooBig,
			"unlocking script size %d is larger than the max allowed size %d",
			len(*uscript),
			t.cfg.MaxScriptSize(),
		)
	}
	if len(*lscript) > t.cfg.MaxScriptSize() {
		return errs.NewError(
			errs.ErrScriptTooBig,
			"locking script size %d is larger than the max allowed size %d",
			len(*uscript),
			t.cfg.MaxScriptSize(),
		)
	}

	// The engine stores the scripts in parsed form using a slice.  This
	// allows multiple scripts to be executed in sequence.  For example,
	// with a pay-to-script-hash transaction, there will be ultimately be
	// a third script to execute.
	t.scripts = make([]ParsedScript, 2)
	for i, script := range []*bscript.Script{uscript, lscript} {
		pscript, err := t.scriptParser.Parse(script)
		if err != nil {
			return err
		}

		t.scripts[i] = pscript
	}

	// The signature script must only contain data pushes when the
	// associated flag is set.
	if t.hasFlag(scriptflag.VerifySigPushOnly) && !t.scripts[0].IsPushOnly() {
		return errs.NewError(errs.ErrNotPushOnly, "signature script is not push only")
	}

	// Advance the program counter to the public key script if the signature
	// script is empty since there is nothing to execute for it in that
	// case.
	if len(*uscript) == 0 {
		t.scriptIdx++
	}

	if t.hasFlag(scriptflag.Bip16) && lscript.IsP2SH() {
		// Only accept input scripts that push data for P2SH.
		if !t.scripts[0].IsPushOnly() {
			return errs.NewError(errs.ErrNotPushOnly, "pay to script hash is not push only")
		}
		t.bip16 = true
	}

	t.dstack = newStack(t.cfg, t.hasFlag(scriptflag.VerifyMinimalData))
	t.astack = newStack(t.cfg, t.hasFlag(scriptflag.VerifyMinimalData))

	if t.tx != nil {
		t.tx.InputIdx(t.inputIdx).PreviousTxScript = t.prevOutput.LockingScript
		t.tx.InputIdx(t.inputIdx).PreviousTxSatoshis = t.prevOutput.Satoshis
	}

	t.state = t
	if opts.debugger == nil {
		opts.debugger = &nopDebugger{}
		t.state = &nopStateHandler{}
	}
	t.debug = opts.debugger
	t.dstack.debug = t.debug
	t.dstack.sh = t.state
	t.astack.debug = t.debug
	t.astack.sh = t.state

	if opts.state != nil {
		t.SetState(opts.state)
	}

	return nil
}

func (t *thread) execute() error {
	if err := func() error {
		defer t.afterExecute()
		t.beforeExecute()
		for {
			t.beforeStep()

			done, err := t.Step()
			if err != nil {
				return err
			}

			t.afterStep()
			if done {
				return nil
			}
		}
	}(); err != nil {
		return err
	}

	return t.CheckErrorCondition(true)
}
