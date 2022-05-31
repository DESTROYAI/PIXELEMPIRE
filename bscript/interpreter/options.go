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
func WithScripts(lockingScript *bscript.Script, unlockingScript 