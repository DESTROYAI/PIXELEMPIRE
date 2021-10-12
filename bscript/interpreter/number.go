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
// are only allowed to be in the range determined by a maximum number of bytes,
// on a per opcode basis, an error will be returned when the provided bytes
// would result in a number outside that range.  In particular, the range for
// the vast majority of opcodes dealing with numeric values are limited to 4
// bytes and therefore will pass that value to this function resulting in an
// allowed range of [-2^31 + 1, 2^31 - 1].
//
// The requireMinimal flag causes an error to be returned if additional checks
// on the encoding determine it is not represented with the smallest possible
// number of bytes or is the negative 0 encoding, [0x80].  For example, consider
// the number 127.  It could be encoded as [0x7f], [0x7f 0x00],
// [0x7f 0x00 0x00 ...], etc.  All forms except [0x7f] will return an error with
// requireMinimal enabled.
//
// The scriptNumLen is the maximum number of bytes the encoded value can be
// before an errs.ErrStackNumberTooBig is returned.  This effectively limits the
// range of allowed values.
// WARNING:  Great care should be taken if passing a value larger than
// defaultScriptNumLen, which could lead to addition and multiplication
// overflows.
//
// See the Bytes function documentation for example encodings.
func makeScriptNumber(bb []byte, scriptNumLen int, requireMinimal, afterGenesis bool) (*scriptNumber, error) {
	// Interpreting data requires that it is not larger than the passed scriptNumLen value.
	if len(bb) > scriptNumLen {
		return &scriptNumber{val: big.NewInt(0), afterGenesis: false}, errs.NewError(
			errs.ErrNumberTooBig,
			"numeric value encoded as %x is %d bytes which exceeds the max allowed of %d",
			bb, len(bb), scriptNumLen,
		)
	}

	// Enforce minimal encoded if requested.
	if requireMinimal {
		if err := checkMinimalDataEncoding(bb); err != nil {
			return &scriptNumber{
				val:          big.NewInt(0),
				afterGenesis: false,
			}, err
		}
	}

	// Zero is encoded as an empty byte slice.
	if len(bb) == 0 {
		return &scriptNumber{
			afterGenesis: afterGenesis,
			val:          big.NewInt(0),
		}, nil
	}

	// Decode from little endian.
	//
	// The following is the equivalent of:
	//    var v int64
	//    for i, b := range bb {
	//        v |= int64(b) << uint8(8*i)
	//    }
	v := new(big.Int)
	for i, b := range bb {
		v.Or(v, new(big.Int).Lsh(new(big.Int).SetBytes([]byte{b}), uint(8*i)))
	}

	// When the most significant byte of the input bytes has the sign bit
	// set, the result is negative.  So, remove the sign bit from the result
	// and make it negative.
	//
	// The following is the equivalent of:
	//    if bb[len(bb)-1]&0x80 != 0 {
	//        v &= ^(int64(0x80) << uint8(8*(len(bb)-1)))
	//        return -v, nil
	//    }
	if bb[len(bb)-1]&0x80 != 0 {
		// The maximum length of bb has already been determined to be 4
		// above, so uint8 is enough to cover the max 