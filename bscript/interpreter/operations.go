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
	bscript.OpDATA23:    {bscript.OpDATA23, "OP_DATA_23", 24, opcodePushData},
	bscript.OpDATA24:    {bscript.OpDATA24, "OP_DATA_24", 25, opcodePushData},
	bscript.OpDATA25:    {bscript.OpDATA25, "OP_DATA_25", 26, opcodePushData},
	bscript.OpDATA26:    {bscript.OpDATA26, "OP_DATA_26", 27, opcodePushData},
	bscript.OpDATA27:    {bscript.OpDATA27, "OP_DATA_27", 28, opcodePushData},
	bscript.OpDATA28:    {bscript.OpDATA28, "OP_DATA_28", 29, opcodePushData},
	bscript.OpDATA29:    {bscript.OpDATA29, "OP_DATA_29", 30, opcodePushData},
	bscript.OpDATA30:    {bscript.OpDATA30, "OP_DATA_30", 31, opcodePushData},
	bscript.OpDATA31:    {bscript.OpDATA31, "OP_DATA_31", 32, opcodePushData},
	bscript.OpDATA32:    {bscript.OpDATA32, "OP_DATA_32", 33, opcodePushData},
	bscript.OpDATA33:    {bscript.OpDATA33, "OP_DATA_33", 34, opcodePushData},
	bscript.OpDATA34:    {bscript.OpDATA34, "OP_DATA_34", 35, opcodePushData},
	bscript.OpDATA35:    {bscript.OpDATA35, "OP_DATA_35", 36, opcodePushData},
	bscript.OpDATA36:    {bscript.OpDATA36, "OP_DATA_36", 37, opcodePushData},
	bscript.OpDATA37:    {bscript.OpDATA37, "OP_DATA_37", 38, opcodePushData},
	bscript.OpDATA38:    {bscript.OpDATA38, "OP_DATA_38", 39, opcodePushData},
	bscript.OpDATA39:    {bscript.OpDATA39, "OP_DATA_39", 40, opcodePushData},
	bscript.OpDATA40:    {bscript.OpDATA40, "OP_DATA_40", 41, opcodePushData},
	bscript.OpDATA41:    {bscript.OpDATA41, "OP_DATA_41", 42, opcodePushData},
	bscript.OpDATA42:    {bscript.OpDATA42, "OP_DATA_42", 43, opcodePushData},
	bscript.OpDATA43:    {bscript.OpDATA43, "OP_DATA_43", 44, opcodePushData},
	bscript.OpDATA44:    {bscript.OpDATA44, "OP_DATA_44", 45, opcodePushData},
	bscript.OpDATA45:    {bscript.OpDATA45, "OP_DATA_45", 46, opcodePushData},
	bscript.OpDATA46:    {bscript.OpDATA46, "OP_DATA_46", 47, opcodePushData},
	bscript.OpDATA47:    {bscript.OpDATA47, "OP_DATA_47", 48, opcodePushData},
	bscript.OpDATA48:    {bscript.OpDATA48, "OP_DATA_48", 49, opcodePushData},
	bscript.OpDATA49:    {bscript.OpDATA49, "OP_DATA_49", 50, opcodePushData},
	bscript.OpDATA50:    {bscript.OpDATA50, "OP_DATA_50", 51, opcodePushData},
	bscript.OpDATA51:    {bscript.OpDATA51, "OP_DATA_51", 52, opcodePushData},
	bscript.OpDATA52:    {bscript.OpDATA52, "OP_DATA_52", 53, opcodePushData},
	bscript.OpDATA53:    {bscript.OpDATA53, "OP_DATA_53", 54, opcodePushData},
	bscript.OpDATA54:    {bscript.OpDATA54, "OP_DATA_54", 55, opcodePushData},
	bscript.OpDATA55:    {bscript.OpDATA55, "OP_DATA_55", 56, opcodePushData},
	bscript.OpDATA56:    {bscript.OpDATA56, "OP_DATA_56", 57, opcodePushData},
	bscript.OpDATA57:    {bscript.OpDATA57, "OP_DATA_57", 58, opcodePushData},
	bscript.OpDATA58:    {bscript.OpDATA58, "OP_DATA_58", 59, opcodePushData},
	bscript.OpDATA59:    {bscript.OpDATA59, "OP_DATA_59", 60, opcodePushData},
	bscript.OpDATA60:    {bscript.OpDATA60, "OP_DATA_60", 61, opcodePushData},
	bscript.OpDATA61:    {bscript.OpDATA61, "OP_DATA_61", 62, opcodePushData},
	bscript.OpDATA62:    {bscript.OpDATA62, "OP_DATA_62", 63, opcodePushData},
	bscript.OpDATA63:    {bscript.OpDATA63, "OP_DATA_63", 64, opcodePushData},
	bscript.OpDATA64:    {bscript.OpDATA64, "OP_DATA_64", 65, opcodePushData},
	bscript.OpDATA65:    {bscript.OpDATA65, "OP_DATA_65", 66, opcodePushData},
	bscript.OpDATA66:    {bscript.OpDATA66, "OP_DATA_66", 67, opcodePushData},
	bscript.OpDATA67:    {bscript.OpDATA67, "OP_DATA_67", 68, opcodePushData},
	bscript.OpDATA68:    {bscript.OpDATA68, "OP_DATA_68", 69, opcodePushData},
	bscript.OpDATA69:    {bscript.OpDATA69, "OP_DATA_69", 70, opcodePushData},
	bscript.OpDATA70:    {bscript.OpDATA70, "OP_DATA_70", 71, opcodePushData},
	bscript.OpDATA71:    {bscript.OpDATA71, "OP_DATA_71", 72, opcodePushData},
	bscript.OpDATA72:    {bscript.OpDATA72, "OP_DATA_72", 73, opcodePushData},
	bscript.OpDATA73:    {bscript.OpDATA73, "OP_DATA_73", 74, opcodePushData},
	bscript.OpDATA74:    {bscript.OpDATA74, "OP_DATA_74", 75, opcodePushData},
	bscript.OpDATA75:    {bscript.OpDATA75, "OP_DATA_75", 76, opcodePushData},
	bscript.OpPUSHDATA1: {bscript.OpPUSHDATA1, "OP_PUSHDATA1", -1, opcodePushData},
	bscript.OpPUSHDATA2: {bscript.OpPUSHDATA2, "OP_PUSHDATA2", -2, opcodePushData},
	bscript.OpPUSHDATA4: {bscript.OpPUSHDATA4, "OP_PUSHDATA4", -4, opcodePushData},
	bscript.Op1NEGATE:   {bscript.Op1NEGATE, "OP_1NEGATE", 1, opcode1Negate},
	bscript.OpRESERVED:  {bscript.OpRESERVED, "OP_RESERVED", 1, opcodeReserved},
	bscript.OpTRUE:      {bscript.OpTRUE, "OP_1", 1, opcodeN},
	bscript.Op2:         {bscript.Op2, "OP_2", 1, opcodeN},
	bscript.Op3:         {bscript.Op3, "OP_3", 1, opcodeN},
	bscript.Op4:         {bscript.Op4, "OP_4", 1, opcodeN},
	bscript.Op5:         {bscript.Op5, "OP_5", 1, opcodeN},
	bscript.Op6:         {bscript.Op6, "OP_6", 1, opcodeN},
	bscript.Op7:         {bscript.Op7, "OP_7", 1, opcodeN},
	bscript.Op8:         {bscript.Op8, "OP_8", 1, opcodeN},
	bscript.Op9:         {bscript.Op9, "OP_9", 1, opcodeN},
	bscript.Op10:        {bscript.Op10, "OP_10", 1, opcodeN},
	bscript.Op11:        {bscript.Op11, "OP_11", 1, opcodeN},
	bscript.Op12:        {bscript.Op12, "OP_12", 1, opcodeN},
	bscript.Op13:        {bscript.Op13, "OP_13", 1, opcodeN},
	bscript.Op14:        {bscript.Op14, "OP_14", 1, opcodeN},
	bscript.Op15:        {bscript.Op15, "OP_15", 1, opcodeN},
	bscript.Op16:        {bscript.Op16, "OP_16", 1, opcodeN},

	// Control opcodes.
	bscript.OpNOP:                 {bscript.OpNOP, "OP_NOP", 1, opcodeNop},
	bscript.OpVER:                 {bscript.OpVER, "OP_VER", 1, opcodeReserved},
	bscript.OpIF:                  {bscript.OpIF, "OP_IF", 1, opcodeIf},
	bscript.OpNOTIF:               {bscript.OpNOTIF, "OP_NOTIF", 1, opcodeNotIf},
	bscript.OpVERIF:               {bscript.OpVERIF, "OP_VERIF", 1, opcodeVerConditional},
	bscript.OpVERNOTIF:            {bscript.OpVERNOTIF, "OP_VERNOTIF", 1, opcodeVerConditional},
	bscript.OpELSE:                {bscript.OpELSE, "OP_ELSE", 1, opcodeElse},
	bscript.OpENDIF:               {bscript.OpENDIF, "OP_ENDIF", 1, opcodeEndif},
	bscript.OpVERIFY:              {bscript.OpVERIFY, "OP_VERIFY", 1, opcodeVerify},
	bscript.OpRETURN:              {bscript.OpRETURN, "OP_RETURN", 1, opcodeReturn},
	bscript.OpCHECKLOCKTIMEVERIFY: {bscript.OpCHECKLOCKTIMEVERIFY, "OP_CHECKLOCKTIMEVERIFY", 1, opcodeCheckLockTimeVerify},
	bscript.OpCHECKSEQUENCEVERIFY: {bscript.OpCHECKSEQUENCEVERIFY, "OP_CHECKSEQUENCEVERIFY", 1, opcodeCheckSequenceVerify},

	// Stack opcodes.
	bscript.OpTOALTSTACK:   {bscript.OpTOALTSTACK, "OP_TOALTSTACK", 1, opcodeToAltStack},
	bscript.OpFROMALTSTACK: {bscript.OpFROMALTSTACK, "OP_FROMALTSTACK", 1, opcodeFromAltStack},
	bscript.Op2DROP:        {bscript.Op2DROP, "OP_2DROP", 1, opcode2Drop},
	bscript.Op2DUP:         {bscript.Op2DUP, "OP_2DUP", 1, opcode2Dup},
	bscript.Op3DUP:         {bscript.Op3DUP, "OP_3DUP", 1, opcode3Dup},
	bscript.Op2OVER:        {bs