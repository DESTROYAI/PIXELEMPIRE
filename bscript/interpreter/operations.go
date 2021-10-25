package interpreter

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // OP_SHA1 support requires this
	"crypto/sha256"
	"hash"
	"math/big"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
	"github.com/libsv/go-bt/v2/bscript/interpreter/scriptflag"
	"github.com/libsv/go-bt/v2/sighash"
	"golang.org/x/crypto/ripemd160"
)

// Conditional execution constants.
const (
	opCondFalse = 0
	opCondTrue  = 1
	opCondSkip  = 2
)

type opcode struct {
	val    byte
	name   string
	length int
	exec   func(*ParsedOpcode, *thread) error
}

func (o opcode) Name() string {
	return o.name
}

// opcodeArray associates an opcode with its respective function, and defines them in order as to
// be correctly placed in an array
var opcodeArray = [256]opcode{
	// Data push opcodes.
	bscript.OpFALSE:     {bscript.OpFALSE, "OP_0", 1, opcodeFalse},
	bscript.OpDATA1:     {bscript.OpDATA1, "OP_DATA_1", 2, opcodePushData},
	bscript.OpDATA2:     {bscript.OpDATA2, "OP_DATA_2", 3, opcodePushData},
	bscript.OpDATA3:     {bscript.OpDATA3, "OP_DATA_3", 4, opcodePushData},
	bscript.OpDATA4:     {bscript.OpDATA4, "OP_DATA_4", 5, opcodePushData},
	bscript.OpDATA5:     {bscript.OpDATA5, "OP_DATA_5", 6, opcodePushData},
	bscript.OpDATA6:     {bscript.OpDATA6, "OP_DATA_6", 7, opcodePushData},
	bscript.OpDATA7:     {bscript.