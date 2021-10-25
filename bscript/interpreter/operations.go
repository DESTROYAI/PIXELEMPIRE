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
	bscript.OpDATA7:     {bscript.OpDATA7, "OP_DATA_7", 8, opcodePushData},
	bscript.OpDATA8:     {bscript.OpDATA8, "OP_DATA_8", 9, opcodePushData},
	bscript.OpDATA9:     {bscript.OpDATA9, "OP_DATA_9", 10, opcodePushData},
	bscript.OpDATA10:    {bscript.OpDATA10, "OP_DATA_10", 11, opcodePushData},
	bscript.OpDATA11:    {bscript.OpDATA11, "OP_DATA_11", 12, opcodePushData},
	bscript.OpDATA12:    {bscript.OpDATA12, "OP_DATA_12", 13, opcodePushData},
	bscript.OpDATA13:    {bscript.OpDATA13, "OP_DATA_13", 14, opcodePushData},
	bscript.OpDATA14:    {bscript.OpDATA14, "OP_DATA_14", 15, opcodePushData},
	bscript.OpDATA15:    {bscript.OpDATA15, "OP_DATA_15", 16, opcodePushData},
	bscript.OpDATA16:    {bscript.OpDATA16, "OP_DATA_16", 17, opcodePushData},
	bscript.OpDATA17:    {bscript.OpDATA17, "OP_DATA_17", 18, opcodePushData},
	bscript.OpDATA18:    {bscript.OpDATA18, "OP_DATA_18", 19, opcodePushData},
	bscript.OpDATA19:    {bscript.OpDATA19, "OP_DATA_19", 20, opcodePushData},
	bscript.OpDATA20:    {bscript.OpDATA20, "OP_DATA_20", 21, opcodePushData},
	bscript.OpDATA21:    {bscript.OpDATA21, "OP_DATA_21", 22, opcodePushData},
	bscript.OpDATA22:    {bscript.OpDATA22, "OP_DATA_22", 23, opcodePushData},
	