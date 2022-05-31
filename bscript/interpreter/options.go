package interpreter

import (
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
)

// ExecutionOptionFunc for setting execution options.
type ExecutionOptionFunc func(p *execOpts)

// WithTx configure the execution to run again a tx.
func WithTx(tx *bt.Tx, inputIdx int, prevOutput *bt.Output) ExecutionOptionFunc {
	return func(p *execOpts) {
		p.tx = tx
		p.previousTxOut = prevOutput
		p.inputIdx = inputIdx
	}
}

// WithScripts configure the execution to run again a set of *bscript.Script.
func WithScripts(lockingScript *bscript.Script, unlockingScript *bscript.Script) ExecutionOptionFunc {
	return func(p *execOpts) {
		p.lockingScript = lockingScript
		p.unlockingScript = unlockingScript
	}
}

// WithAfterGenesis configure the execution to operate in an after-genesis context.
func WithAfterGenesis() ExecutionOptionFunc {
	return func(p *execOpts) {
		p.flags.AddFlag(scriptflag.UTXOAfterGenesis)
	}
}

// WithForkID configure the execution to allow a tx with a fork id.
func WithForkID() ExecutionOptionFunc {
	return func(p *execOpts) {
		p.flags.AddFlag(scriptflag.EnableSighashForkID)
	}
}

// WithP2SH