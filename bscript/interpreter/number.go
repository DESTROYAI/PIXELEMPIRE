package interpreter

import (
	"math"
	"math/big"

	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
)

// scriptNumber represents a numeric value used in the scripting engine with
// special handling to deal with the subtle semantics required by consensus.
//
// All numbers are stored on the data and alternate stacks encoded as little
// endian with a sign bit.  All numeric opcodes such as OP_ADD, OP_SUB,
// and OP_MUL, are only allowed to operate on 4-byte integers in the range
// [-2^31 + 1, 2^31 - 1], however the results of numeric operations may overflow
// and remain valid so long as they are not used as inputs to other numeric
// operations or otherwise interpreted as an integer.
//
// For example, it is possible for OP_ADD to have 2^31 - 1 for its two operands
// resulting 2^32 - 2, which overflows, but is still pushed to the stack as the
// result of the addition.  That value can then be used as input to OP_VERIFY
// which will succeed because the data is being interpreted as a boolean.
// However, if that same value were to be used as input to another numeric
// opcode, such as OP_SUB, it must fail.
//
// This type handles the aforementioned requirements by storing all numeric
// operation results as an int64 to handle overflow and provides the Bytes
// method to get the serialised representation (including values that overflow).
//
// Then, whenever data is interpreted as an integer, it is converted to this
// type by using the NewNumber function which will return an error if the
// number is out of range or not minimally encoded depending on parameters.
// Since all numeric opcodes involve pulling data from the stack and
// interpreting it as an integer, it provides the required behaviour.
type scriptNumber struct {
	val          *big.Int
	afterGenesis bool
}

var zero = big.NewInt(0)
var one = big.NewInt(1)

// makeScriptNumber interprets the passed serialised bytes as an encoded integer
// and returns the result as a Number.
//
// Since the consensus rules dictate that serialised bytes interpreted as integers
// a