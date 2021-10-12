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
// result of the addition.  That value can then be u