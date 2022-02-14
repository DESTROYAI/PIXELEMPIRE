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
	bscript.Op2OVER:        {bscript.Op2OVER, "OP_2OVER", 1, opcode2Over},
	bscript.Op2ROT:         {bscript.Op2ROT, "OP_2ROT", 1, opcode2Rot},
	bscript.Op2SWAP:        {bscript.Op2SWAP, "OP_2SWAP", 1, opcode2Swap},
	bscript.OpIFDUP:        {bscript.OpIFDUP, "OP_IFDUP", 1, opcodeIfDup},
	bscript.OpDEPTH:        {bscript.OpDEPTH, "OP_DEPTH", 1, opcodeDepth},
	bscript.OpDROP:         {bscript.OpDROP, "OP_DROP", 1, opcodeDrop},
	bscript.OpDUP:          {bscript.OpDUP, "OP_DUP", 1, opcodeDup},
	bscript.OpNIP:          {bscript.OpNIP, "OP_NIP", 1, opcodeNip},
	bscript.OpOVER:         {bscript.OpOVER, "OP_OVER", 1, opcodeOver},
	bscript.OpPICK:         {bscript.OpPICK, "OP_PICK", 1, opcodePick},
	bscript.OpROLL:         {bscript.OpROLL, "OP_ROLL", 1, opcodeRoll},
	bscript.OpROT:          {bscript.OpROT, "OP_ROT", 1, opcodeRot},
	bscript.OpSWAP:         {bscript.OpSWAP, "OP_SWAP", 1, opcodeSwap},
	bscript.OpTUCK:         {bscript.OpTUCK, "OP_TUCK", 1, opcodeTuck},

	// Splice opcodes.
	bscript.OpCAT:     {bscript.OpCAT, "OP_CAT", 1, opcodeCat},
	bscript.OpSPLIT:   {bscript.OpSPLIT, "OP_SPLIT", 1, opcodeSplit},
	bscript.OpNUM2BIN: {bscript.OpNUM2BIN, "OP_NUM2BIN", 1, opcodeNum2bin},
	bscript.OpBIN2NUM: {bscript.OpBIN2NUM, "OP_BIN2NUM", 1, opcodeBin2num},
	bscript.OpSIZE:    {bscript.OpSIZE, "OP_SIZE", 1, opcodeSize},

	// Bitwise logic opcodes.
	bscript.OpINVERT:      {bscript.OpINVERT, "OP_INVERT", 1, opcodeInvert},
	bscript.OpAND:         {bscript.OpAND, "OP_AND", 1, opcodeAnd},
	bscript.OpOR:          {bscript.OpOR, "OP_OR", 1, opcodeOr},
	bscript.OpXOR:         {bscript.OpXOR, "OP_XOR", 1, opcodeXor},
	bscript.OpEQUAL:       {bscript.OpEQUAL, "OP_EQUAL", 1, opcodeEqual},
	bscript.OpEQUALVERIFY: {bscript.OpEQUALVERIFY, "OP_EQUALVERIFY", 1, opcodeEqualVerify},
	bscript.OpRESERVED1:   {bscript.OpRESERVED1, "OP_RESERVED1", 1, opcodeReserved},
	bscript.OpRESERVED2:   {bscript.OpRESERVED2, "OP_RESERVED2", 1, opcodeReserved},

	// Numeric related opcodes.
	bscript.Op1ADD:               {bscript.Op1ADD, "OP_1ADD", 1, opcode1Add},
	bscript.Op1SUB:               {bscript.Op1SUB, "OP_1SUB", 1, opcode1Sub},
	bscript.Op2MUL:               {bscript.Op2MUL, "OP_2MUL", 1, opcodeDisabled},
	bscript.Op2DIV:               {bscript.Op2DIV, "OP_2DIV", 1, opcodeDisabled},
	bscript.OpNEGATE:             {bscript.OpNEGATE, "OP_NEGATE", 1, opcodeNegate},
	bscript.OpABS:                {bscript.OpABS, "OP_ABS", 1, opcodeAbs},
	bscript.OpNOT:                {bscript.OpNOT, "OP_NOT", 1, opcodeNot},
	bscript.Op0NOTEQUAL:          {bscript.Op0NOTEQUAL, "OP_0NOTEQUAL", 1, opcode0NotEqual},
	bscript.OpADD:                {bscript.OpADD, "OP_ADD", 1, opcodeAdd},
	bscript.OpSUB:                {bscript.OpSUB, "OP_SUB", 1, opcodeSub},
	bscript.OpMUL:                {bscript.OpMUL, "OP_MUL", 1, opcodeMul},
	bscript.OpDIV:                {bscript.OpDIV, "OP_DIV", 1, opcodeDiv},
	bscript.OpMOD:                {bscript.OpMOD, "OP_MOD", 1, opcodeMod},
	bscript.OpLSHIFT:             {bscript.OpLSHIFT, "OP_LSHIFT", 1, opcodeLShift},
	bscript.OpRSHIFT:             {bscript.OpRSHIFT, "OP_RSHIFT", 1, opcodeRShift},
	bscript.OpBOOLAND:            {bscript.OpBOOLAND, "OP_BOOLAND", 1, opcodeBoolAnd},
	bscript.OpBOOLOR:             {bscript.OpBOOLOR, "OP_BOOLOR", 1, opcodeBoolOr},
	bscript.OpNUMEQUAL:           {bscript.OpNUMEQUAL, "OP_NUMEQUAL", 1, opcodeNumEqual},
	bscript.OpNUMEQUALVERIFY:     {bscript.OpNUMEQUALVERIFY, "OP_NUMEQUALVERIFY", 1, opcodeNumEqualVerify},
	bscript.OpNUMNOTEQUAL:        {bscript.OpNUMNOTEQUAL, "OP_NUMNOTEQUAL", 1, opcodeNumNotEqual},
	bscript.OpLESSTHAN:           {bscript.OpLESSTHAN, "OP_LESSTHAN", 1, opcodeLessThan},
	bscript.OpGREATERTHAN:        {bscript.OpGREATERTHAN, "OP_GREATERTHAN", 1, opcodeGreaterThan},
	bscript.OpLESSTHANOREQUAL:    {bscript.OpLESSTHANOREQUAL, "OP_LESSTHANOREQUAL", 1, opcodeLessThanOrEqual},
	bscript.OpGREATERTHANOREQUAL: {bscript.OpGREATERTHANOREQUAL, "OP_GREATERTHANOREQUAL", 1, opcodeGreaterThanOrEqual},
	bscript.OpMIN:                {bscript.OpMIN, "OP_MIN", 1, opcodeMin},
	bscript.OpMAX:                {bscript.OpMAX, "OP_MAX", 1, opcodeMax},
	bscript.OpWITHIN:             {bscript.OpWITHIN, "OP_WITHIN", 1, opcodeWithin},

	// Crypto opcodes.
	bscript.OpRIPEMD160:           {bscript.OpRIPEMD160, "OP_RIPEMD160", 1, opcodeRipemd160},
	bscript.OpSHA1:                {bscript.OpSHA1, "OP_SHA1", 1, opcodeSha1},
	bscript.OpSHA256:              {bscript.OpSHA256, "OP_SHA256", 1, opcodeSha256},
	bscript.OpHASH160:             {bscript.OpHASH160, "OP_HASH160", 1, opcodeHash160},
	bscript.OpHASH256:             {bscript.OpHASH256, "OP_HASH256", 1, opcodeHash256},
	bscript.OpCODESEPARATOR:       {bscript.OpCODESEPARATOR, "OP_CODESEPARATOR", 1, opcodeCodeSeparator},
	bscript.OpCHECKSIG:            {bscript.OpCHECKSIG, "OP_CHECKSIG", 1, opcodeCheckSig},
	bscript.OpCHECKSIGVERIFY:      {bscript.OpCHECKSIGVERIFY, "OP_CHECKSIGVERIFY", 1, opcodeCheckSigVerify},
	bscript.OpCHECKMULTISIG:       {bscript.OpCHECKMULTISIG, "OP_CHECKMULTISIG", 1, opcodeCheckMultiSig},
	bscript.OpCHECKMULTISIGVERIFY: {bscript.OpCHECKMULTISIGVERIFY, "OP_CHECKMULTISIGVERIFY", 1, opcodeCheckMultiSigVerify},

	// Reserved opcodes.
	bscript.OpNOP1:  {bscript.OpNOP1, "OP_NOP1", 1, opcodeNop},
	bscript.OpNOP4:  {bscript.OpNOP4, "OP_NOP4", 1, opcodeNop},
	bscript.OpNOP5:  {bscript.OpNOP5, "OP_NOP5", 1, opcodeNop},
	bscript.OpNOP6:  {bscript.OpNOP6, "OP_NOP6", 1, opcodeNop},
	bscript.OpNOP7:  {bscript.OpNOP7, "OP_NOP7", 1, opcodeNop},
	bscript.OpNOP8:  {bscript.OpNOP8, "OP_NOP8", 1, opcodeNop},
	bscript.OpNOP9:  {bscript.OpNOP9, "OP_NOP9", 1, opcodeNop},
	bscript.OpNOP10: {bscript.OpNOP10, "OP_NOP10", 1, opcodeNop},

	// Undefined opcodes.
	bscript.OpUNKNOWN186: {bscript.OpUNKNOWN186, "OP_UNKNOWN186", 1, opcodeInvalid},
	bscript.OpUNKNOWN187: {bscript.OpUNKNOWN187, "OP_UNKNOWN187", 1, opcodeInvalid},
	bscript.OpUNKNOWN188: {bscript.OpUNKNOWN188, "OP_UNKNOWN188", 1, opcodeInvalid},
	bscript.OpUNKNOWN189: {bscript.OpUNKNOWN189, "OP_UNKNOWN189", 1, opcodeInvalid},
	bscript.OpUNKNOWN190: {bscript.OpUNKNOWN190, "OP_UNKNOWN190", 1, opcodeInvalid},
	bscript.OpUNKNOWN191: {bscript.OpUNKNOWN191, "OP_UNKNOWN191", 1, opcodeInvalid},
	bscript.OpUNKNOWN192: {bscript.OpUNKNOWN192, "OP_UNKNOWN192", 1, opcodeInvalid},
	bscript.OpUNKNOWN193: {bscript.OpUNKNOWN193, "OP_UNKNOWN193", 1, opcodeInvalid},
	bscript.OpUNKNOWN194: {bscript.OpUNKNOWN194, "OP_UNKNOWN194", 1, opcodeInvalid},
	bscript.OpUNKNOWN195: {bscript.OpUNKNOWN195, "OP_UNKNOWN195", 1, opcodeInvalid},
	bscript.OpUNKNOWN196: {bscript.OpUNKNOWN196, "OP_UNKNOWN196", 1, opcodeInvalid},
	bscript.OpUNKNOWN197: {bscript.OpUNKNOWN197, "OP_UNKNOWN197", 1, opcodeInvalid},
	bscript.OpUNKNOWN198: {bscript.OpUNKNOWN198, "OP_UNKNOWN198", 1, opcodeInvalid},
	bscript.OpUNKNOWN199: {bscript.OpUNKNOWN199, "OP_UNKNOWN199", 1, opcodeInvalid},
	bscript.OpUNKNOWN200: {bscript.OpUNKNOWN200, "OP_UNKNOWN200", 1, opcodeInvalid},
	bscript.OpUNKNOWN201: {bscript.OpUNKNOWN201, "OP_UNKNOWN201", 1, opcodeInvalid},
	bscript.OpUNKNOWN202: {bscript.OpUNKNOWN202, "OP_UNKNOWN202", 1, opcodeInvalid},
	bscript.OpUNKNOWN203: {bscript.OpUNKNOWN203, "OP_UNKNOWN203", 1, opcodeInvalid},
	bscript.OpUNKNOWN204: {bscript.OpUNKNOWN204, "OP_UNKNOWN204", 1, opcodeInvalid},
	bscript.OpUNKNOWN205: {bscript.OpUNKNOWN205, "OP_UNKNOWN205", 1, opcodeInvalid},
	bscript.OpUNKNOWN206: {bscript.OpUNKNOWN206, "OP_UNKNOWN206", 1, opcodeInvalid},
	bscript.OpUNKNOWN207: {bscript.OpUNKNOWN207, "OP_UNKNOWN207", 1, opcodeInvalid},
	bscript.OpUNKNOWN208: {bscript.OpUNKNOWN208, "OP_UNKNOWN208", 1, opcodeInvalid},
	bscript.OpUNKNOWN209: {bscript.OpUNKNOWN209, "OP_UNKNOWN209", 1, opcodeInvalid},
	bscript.OpUNKNOWN210: {bscript.OpUNKNOWN210, "OP_UNKNOWN210", 1, opcodeInvalid},
	bscript.OpUNKNOWN211: {bscript.OpUNKNOWN211, "OP_UNKNOWN211", 1, opcodeInvalid},
	bscript.OpUNKNOWN212: {bscript.OpUNKNOWN212, "OP_UNKNOWN212", 1, opcodeInvalid},
	bscript.OpUNKNOWN213: {bscript.OpUNKNOWN213, "OP_UNKNOWN213", 1, opcodeInvalid},
	bscript.OpUNKNOWN214: {bscript.OpUNKNOWN214, "OP_UNKNOWN214", 1, opcodeInvalid},
	bscript.OpUNKNOWN215: {bscript.OpUNKNOWN215, "OP_UNKNOWN215", 1, opcodeInvalid},
	bscript.OpUNKNOWN216: {bscript.OpUNKNOWN216, "OP_UNKNOWN216", 1, opcodeInvalid},
	bscript.OpUNKNOWN217: {bscript.OpUNKNOWN217, "OP_UNKNOWN217", 1, opcodeInvalid},
	bscript.OpUNKNOWN218: {bscript.OpUNKNOWN218, "OP_UNKNOWN218", 1, opcodeInvalid},
	bscript.OpUNKNOWN219: {bscript.OpUNKNOWN219, "OP_UNKNOWN219", 1, opcodeInvalid},
	bscript.OpUNKNOWN220: {bscript.OpUNKNOWN220, "OP_UNKNOWN220", 1, opcodeInvalid},
	bscript.OpUNKNOWN221: {bscript.OpUNKNOWN221, "OP_UNKNOWN221", 1, opcodeInvalid},
	bscript.OpUNKNOWN222: {bscript.OpUNKNOWN222, "OP_UNKNOWN222", 1, opcodeInvalid},
	bscript.OpUNKNOWN223: {bscript.OpUNKNOWN223, "OP_UNKNOWN223", 1, opcodeInvalid},
	bscript.OpUNKNOWN224: {bscript.OpUNKNOWN224, "OP_UNKNOWN224", 1, opcodeInvalid},
	bscript.OpUNKNOWN225: {bscript.OpUNKNOWN225, "OP_UNKNOWN225", 1, opcodeInvalid},
	bscript.OpUNKNOWN226: {bscript.OpUNKNOWN226, "OP_UNKNOWN226", 1, opcodeInvalid},
	bscript.OpUNKNOWN227: {bscript.OpUNKNOWN227, "OP_UNKNOWN227", 1, opcodeInvalid},
	bscript.OpUNKNOWN228: {bscript.OpUNKNOWN228, "OP_UNKNOWN228", 1, opcodeInvalid},
	bscript.OpUNKNOWN229: {bscript.OpUNKNOWN229, "OP_UNKNOWN229", 1, opcodeInvalid},
	bscript.OpUNKNOWN230: {bscript.OpUNKNOWN230, "OP_UNKNOWN230", 1, opcodeInvalid},
	bscript.OpUNKNOWN231: {bscript.OpUNKNOWN231, "OP_UNKNOWN231", 1, opcodeInvalid},
	bscript.OpUNKNOWN232: {bscript.OpUNKNOWN232, "OP_UNKNOWN232", 1, opcodeInvalid},
	bscript.OpUNKNOWN233: {bscript.OpUNKNOWN233, "OP_UNKNOWN233", 1, opcodeInvalid},
	bscript.OpUNKNOWN234: {bscript.OpUNKNOWN234, "OP_UNKNOWN234", 1, opcodeInvalid},
	bscript.OpUNKNOWN235: {bscript.OpUNKNOWN235, "OP_UNKNOWN235", 1, opcodeInvalid},
	bscript.OpUNKNOWN236: {bscript.OpUNKNOWN236, "OP_UNKNOWN236", 1, opcodeInvalid},
	bscript.OpUNKNOWN237: {bscript.OpUNKNOWN237, "OP_UNKNOWN237", 1, opcodeInvalid},
	bscript.OpUNKNOWN238: {bscript.OpUNKNOWN238, "OP_UNKNOWN238", 1, opcodeInvalid},
	bscript.OpUNKNOWN239: {bscript.OpUNKNOWN239, "OP_UNKNOWN239", 1, opcodeInvalid},
	bscript.OpUNKNOWN240: {bscript.OpUNKNOWN240, "OP_UNKNOWN240", 1, opcodeInvalid},
	bscript.OpUNKNOWN241: {bscript.OpUNKNOWN241, "OP_UNKNOWN241", 1, opcodeInvalid},
	bscript.OpUNKNOWN242: {bscript.OpUNKNOWN242, "OP_UNKNOWN242", 1, opcodeInvalid},
	bscript.OpUNKNOWN243: {bscript.OpUNKNOWN243, "OP_UNKNOWN243", 1, opcodeInvalid},
	bscript.OpUNKNOWN244: {bscript.OpUNKNOWN244, "OP_UNKNOWN244", 1, opcodeInvalid},
	bscript.OpUNKNOWN245: {bscript.OpUNKNOWN245, "OP_UNKNOWN245", 1, opcodeInvalid},
	bscript.OpUNKNOWN246: {bscript.OpUNKNOWN246, "OP_UNKNOWN246", 1, opcodeInvalid},
	bscript.OpUNKNOWN247: {bscript.OpUNKNOWN247, "OP_UNKNOWN247", 1, opcodeInvalid},
	bscript.OpUNKNOWN248: {bscript.OpUNKNOWN248, "OP_UNKNOWN248", 1, opcodeInvalid},
	bscript.OpUNKNOWN249: {bscript.OpUNKNOWN249, "OP_UNKNOWN249", 1, opcodeInvalid},

	// Bitcoin Core internal use opcode.  Defined here for completeness.
	bscript.OpSMALLINTEGER: {bscript.OpSMALLINTEGER, "OP_SMALLINTEGER", 1, opcodeInvalid},
	bscript.OpPUBKEYS:      {bscript.OpPUBKEYS, "OP_PUBKEYS", 1, opcodeInvalid},
	bscript.OpUNKNOWN252:   {bscript.OpUNKNOWN252, "OP_UNKNOWN252", 1, opcodeInvalid},
	bscript.OpPUBKEYHASH:   {bscript.OpPUBKEYHASH, "OP_PUBKEYHASH", 1, opcodeInvalid},
	bscript.OpPUBKEY:       {bscript.OpPUBKEY, "OP_PUBKEY", 1, opcodeInvalid},

	bscript.OpINVALIDOPCODE: {bscript.OpINVALIDOPCODE, "OP_INVALIDOPCODE", 1, opcodeInvalid},
}

// *******************************************
// Opcode implementation functions start here.
// *******************************************

// opcodeDisabled is a common handler for disabled opcodes.  It returns an
// appropriate error indicating the opcode is disabled.  While it would
// ordinarily make more sense to detect if the script contains any disabled
// opcodes before executing in an initial parse step, the consensus rules
// dictate the script doesn't fail until the program counter passes over a
// disabled opcode (even when they appear in a branch that is not executed).
func opcodeDisabled(op *ParsedOpcode, t *thread) error {
	return errs.NewError(errs.ErrDisabledOpcode, "attempt to execute disabled opcode %s", op.Name())
}

func opcodeVerConditional(op *ParsedOpcode, t *thread) error {
	if t.afterGenesis && !t.shouldExec(*op) {
		return nil
	}
	return opcodeReserved(op, t)
}

// opcodeReserved is a common handler for all reserved opcodes.  It returns an
// appropriate error indicating the opcode is reserved.
func opcodeReserved(op *ParsedOpcode, t *thread) error {
	return errs.NewError(errs.ErrReservedOpcode, "attempt to execute reserved opcode %s", op.Name())
}

// opcodeInvalid is a common handler for all invalid opcodes.  It returns an
// appropriate error indicating the opcode is invalid.
func opcodeInvalid(op *ParsedOpcode, t *thread) error {
	return errs.NewError(errs.ErrReservedOpcode, "attempt to execute invalid opcode %s", op.Name())
}

// opcodeFalse pushes an empty array to the data stack to represent false.  Note
// that 0, when encoded as a number according to the numeric encoding consensus
// rules, is an empty array.
func opcodeFalse(op *ParsedOpcode, t *thread) error {
	t.dstack.PushByteArray(nil)
	return nil
}

// opcodePushData is a common handler for the vast majority of opcodes that push
// raw data (bytes) to the data stack.
func opcodePushData(op *ParsedOpcode, t *thread) error {
	t.dstack.PushByteArray(op.Data)
	return nil
}

// opcode1Negate pushes -1, encoded as a number, to the data stack.
func opcode1Negate(op *ParsedOpcode, t *thread) error {
	t.dstack.PushInt(&scriptNumber{
		val:          big.NewInt(-1),
		afterGenesis: t.afterGenesis,
	})
	return nil
}

// opcodeN is a common handler for the small integer data push opcodes.  It
// pushes the numeric value the opcode represents (which will be from 1 to 16)
// onto the data stack.
func opcodeN(op *ParsedOpcode, t *thread) error {
	// The opcodes are all defined consecutively, so the numeric value is
	// the difference.
	t.dstack.PushByteArray([]byte{(op.op.val - (bscript.Op1 - 1))})
	return nil
}

// opcodeNop is a common handler for the NOP family of opcodes.  As the name
// implies it generally does nothing, however, it will return an error when
// the flag to discourage use of NOPs is set for select opcodes.
func opcodeNop(op *ParsedOpcode, t *thread) error {
	switch op.op.val {
	case bscript.OpNOP1, bscript.OpNOP4, bscript.OpNOP5,
		bscript.OpNOP6, bscript.OpNOP7, bscript.OpNOP8, bscript.OpNOP9, bscript.OpNOP10:
		if t.hasFlag(scriptflag.DiscourageUpgradableNops) {
			return errs.NewError(
				errs.ErrDiscourageUpgradableNOPs,
				"bscript.OpNOP%d reserved for soft-fork upgrades",
				op.op.val-(bscript.OpNOP1-1),
			)
		}
	}

	return nil
}

// popIfBool pops the top item off the stack and returns a bool
func popIfBool(t *thread) (bool, error) {
	if t.hasFlag(scriptflag.VerifyMinimalIf) {
		b, err := t.dstack.PopByteArray()
		if err != nil {
			return false, err
		}

		if len(b) > 1 {
			return false, errs.NewError(errs.ErrMinimalIf, "conditionl has data of length %d", len(b))
		}
		if len(b) == 1 && b[0] != 1 {
			return false, errs.NewError(errs.ErrMinimalIf, "conditional failed")
		}

		return asBool(b), nil
	}

	return t.dstack.PopBool()
}

// opcodeIf treats the top item on the data stack as a boolean and removes it.
//
// An appropriate entry is added to the conditional stack depending on whether
// the boolean is true and whether this if is on an executing branch in order
// to allow proper execution of further opcodes depending on the conditional
// logic.  When the boolean is true, the first branch will be executed (unless
// this opcode is nested in a non-executed branch).
//
// <expression> if [statements] [else [statements]] endif
//
// Note that, unlike for all non-conditional opcodes, this is executed even when
// it is on a non-executing branch so proper nesting is maintained.
//
// Data stack transformation: [... bool] -> [...]
// Conditional stack transformation: [...] -> [... OpCondValue]
func opcodeIf(op *ParsedOpcode, t *thread) error {
	condVal := opCondFalse
	if t.shouldExec(*op) {
		if t.isBranchExecuting() {
			ok, err := popIfBool(t)
			if err != nil {
				return err
			}

			if ok {
				condVal = opCondTrue
			}
		} else {
			condVal = opCondSkip
		}
	}

	t.condStack = append(t.condStack, condVal)
	t.elseStack.PushBool(false)
	return nil
}

// opcodeNotIf treats the top item on the data stack as a boolean and removes
// it.
//
// An appropriate entry is added to the conditional stack depending on whether
// the boolean is true and whether this if is on an executing branch in order
// to allow proper execution of further opcodes depending on the conditional
// logic.  When the boolean is false, the first branch will be executed (unless
// this opcode is nested in a non-executed branch).
//
// <expression> notif [statements] [else [statements]] endif
//
// Note that, unlike for all non-conditional opcodes, this is executed even when
// it is on a non-executing branch so proper nesting is maintained.
//
// Data stack transformation: [... bool] -> [...]
// Conditional stack transformation: [...] -> [... OpCondValue]
func opcodeNotIf(op *ParsedOpcode, t *thread) error {
	condVal := opCondFalse
	if t.shouldExec(*op) {
		if t.isBranchExecuting() {
			ok, err := popIfBool(t)
			if err != nil {
				return err
			}

			if !ok {
				condVal = opCondTrue
			}
		} else {
			condVal = opCondSkip
		}
	}

	t.condStack = append(t.condStack, condVal)
	t.elseStack.PushBool(false)
	return nil
}

// opcodeElse inverts conditional execution for other half of if/else/endif.
//
// An error is returned if there has not already been a matching bscript.OpIF.
//
// Conditional stack transformation: [... OpCondValue] -> [... !OpCondValue]
func opcodeElse(op *ParsedOpcode, t *thread) error {
	if len(t.condStack) == 0 {
		return errs.NewError(errs.ErrUnbalancedConditional,
			"encountered opcode %s with no matching opcode to begin conditional execution", op.Name())
	}

	// Only one ELSE allowed in IF after genesis
	ok, err := t.elseStack.PopBool()
	if err != nil {
		return err
	}
	if ok {
		return errs.NewError(errs.ErrUnbalancedConditional,
			"encountered opcode %s with no matching opcode to begin conditional execution", op.Name())
	}

	conditionalIdx := len(t.condStack) - 1
	switch t.condStack[conditionalIdx] {
	case opCondTrue:
		t.condStack[conditionalIdx] = opCondFalse
	case opCondFalse:
		t.condStack[conditionalIdx] = opCondTrue
	case opCondSkip:
		// Value doesn't change in skip since it indicates this opcode
		// is nested in a non-executed branch.
	}

	t.elseStack.PushBool(true)
	return nil
}

// opcodeEndif terminates a conditional block, removing the value from the
// conditional execution stack.
//
// An error is returned if there has not already been a matching bscript.OpIF.
//
// Conditional stack transformation: [... OpCondValue] -> [...]
func opcodeEndif(op *ParsedOpcode, t *thread) error {
	if len(t.condStack) == 0 {
		return errs.NewError(errs.ErrUnbalancedConditional,
			"encountered opcode %s with no matching opcode to begin conditional execution", op.Name())
	}

	t.condStack = t.condStack[:len(t.condStack)-1]
	if _, err := t.elseStack.PopBool(); err != nil {
		return err
	}

	return nil
}

// abstractVerify examines the top item on the data stack as a boolean value and
// verifies it evaluates to true.  An error is returned either when there is no
// item on the stack or when that item evaluates to false.  In the latter case
// where the verification fails specifically due to the top item evaluating
// to false, the returned error will use the passed error code.
func abstractVerify(op *ParsedOpcode, t *thread, c errs.ErrorCode) error {
	verified, err := t.dstack.PopBool()
	if err != nil {
		return err
	}
	if !verified {
		return errs.NewError(c, "%s failed", op.Name())
	}

	return nil
}

// opcodeVerify examines the top item on the data stack as a boolean value and
// verifies it evaluates to true.  An error is returned if it does not.
func opcodeVerify(op *ParsedOpcode, t *thread) error {
	return abstractVerify(op, t, errs.ErrVerify)
}

// opcodeReturn returns an appropriate error since it is always an error to
// return early from a script.
func opcodeReturn(op *ParsedOpcode, t *thread) error {
	if !t.afterGenesis {
		return errs.NewError(errs.ErrEarlyReturn, "script returned early")
	}

	t.earlyReturnAfterGenesis = true
	if len(t.condStack) == 0 {
		// Terminate the execution as successful. The remaining of the script does not affect the validity (even in
		// presence of unbalanced IFs, invalid opcodes etc)
		return success()
	}

	return nil
}

// verifyLockTime is a helper function used to validate locktimes.
func verifyLockTime(txLockTime, threshold, lockTime int64) error {
	// The lockTimes in both the script and transaction must be of the same
	// type.
	if !((txLockTime < threshold && lockTime < threshold) ||
		(txLockTime >= threshold && lockTime >= threshold)) {
		return errs.NewError(errs.ErrUnsatisfiedLockTime,
			"mismatched locktime types -- tx locktime %d, stack locktime %d", txLockTime, lockTime)
	}

	if lockTime > txLockTime {
		return errs.NewError(errs.ErrUnsatisfiedLockTime,
			"locktime requirement not satisfied -- locktime is greater than the transaction locktime: %d > %d",
			lockTime, txLockTime)
	}

	return nil
}

// opcodeCheckLockTimeVerify compares the top item on the data stack to the
// LockTime field of the transaction containing the script signature
// validating if the transaction outputs are spendable yet.  If flag
// ScriptVerifyCheckLockTimeVerify is not set, the code continues as if bscript.OpNOP2
// were executed.
func opcodeCheckLockTimeVerify(op *ParsedOpcode, t *thread) error {
	// If the ScriptVerifyCheckLockTimeVerify script flag is not set, treat
	// opcode as bscript.OpNOP2 instead.
	if !t.hasFlag(scriptflag.VerifyCheckLockTimeVerify) || t.afterGenesis {
		if t.hasFlag(scriptflag.DiscourageUpgradableNops) {
			return errs.NewError(errs.ErrDiscourageUpgradableNOPs, "bscript.OpNOP2 reserved for soft-fork upgrades")
		}

		return nil
	}

	// The current transaction locktime is a uint32 resulting in a maximum
	// locktime of 2^32-1 (the year 2106).  However, scriptNums are signed
	// and therefore a standard 4-byte scriptNum would only support up to a
	// maximum of 2^31-1 (the year 2038).  Thus, a 5-byte scriptNum is used
	// here since it will support up to 2^39-1 which allows dates beyond the
	// current locktime limit.
	//
	// PeekByteArray is used here instead of PeekInt because we do not want
	// to be limited to a 4-byte integer for reasons specified above.
	so, err := t.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}
	lockTime, err := makeScriptNumber(so, 5, t.dstack.verifyMinimalData, t.afterGenesis)
	if err != nil {
		return err
	}

	// In the rare event that the argument needs to be < 0 due to some
	// arithmetic being done first, you can always use
	// 0 bscript.OpMAX bscript.OpCHECKLOCKTIMEVERIFY.
	if lockTime.LessThanInt(0) {
		return errs.NewError(errs.ErrNegativeLockTime, "negative lock time: %d", lockTime.Int64())
	}

	// The lock time field of a transaction is either a block height at
	// which the transaction is finalised or a timestamp depending on if the
	// value is before the interpreter.LockTimeThreshold.  When it is under the
	// threshold it is a block height.
	if err = verifyLockTime(int64(t.tx.LockTime), LockTimeThreshold, lockTime.Int64()); err != nil {
		return err
	}

	// The lock time feature can also be disabled, thereby bypassing
	// bscript.OpCHECKLOCKTIMEVERIFY, if every transaction input has been finalised by
	// setting its sequence to the maximum value (bt.MaxTxInSequenceNum).  This
	// condition would result in the transaction being allowed into the blockchain
	// making the opcode ineffective.
	//
	// This condition is prevented by enforcing that the input being used by
	// the opcode is unlocked (its sequence number is less than the max
	// value).  This is sufficient to prove correctness without having to
	// check every input.
	//
	// NOTE: This implies that even if the transaction is not finalised due to
	// another input being unlocked, the opcode execution will still fail when the
	// input being used by the opcode is locked.
	if t.tx.Inputs[t.inputIdx].SequenceNumber == bt.MaxTxInSequenceNum {
		return errs.NewError(errs.ErrUnsatisfiedLockTime, "transaction input is finalised")
	}

	return nil
}

// opcodeCheckSequenceVerify compares the top item on the data stack to the
// LockTime field of the transaction containing the script signature
// validating if the transaction outputs are spendable yet.  If flag
// ScriptVerifyCheckSequenceVerify is not set, the code continues as if bscript.OpNOP3
// were executed.
func opcodeCheckSequenceVerify(op *ParsedOpcode, t *thread) error {
	// If the ScriptVerifyCheckSequenceVerify script flag is not set, treat
	// opcode as bscript.OpNOP3 instead.
	if !t.hasFlag(scriptflag.VerifyCheckSequenceVerify) || t.afterGenesis {
		if t.hasFlag(scriptflag.DiscourageUpgradableNops) {
			return errs.NewError(errs.ErrDiscourageUpgradableNOPs, "bscript.OpNOP3 reserved for soft-fork upgrades")
		}

		return nil
	}

	// The current transaction sequence is a uint32 resulting in a maximum
	// sequence of 2^32-1.  However, scriptNums are signed and therefore a
	// standard 4-byte scriptNum would only support up to a maximum of
	// 2^31-1.  Thus, a 5-byte scriptNum is used here since it will support
	// up to 2^39-1 which allows sequences beyond the current sequence
	// limit.
	//
	// PeekByteArray is used here instead of PeekInt because we do not want
	// to be limited to a 4-byte integer for reasons specified above.
	so, err := t.dstack.PeekByteArray(0)
	if err != nil {
		return err
	}
	stackSequence, err := makeScriptNumber(so, 5, t.dstack.verifyMinimalData, t.afterGenesis)
	if err != nil {
		return err
	}

	// In the rare event that the argument needs to be < 0 due to some
	// arithmetic being done first, you can always use
	// 0 bscript.OpMAX bscript.OpCHECKSEQUENCEVERIFY.
	if stackSequence.LessThanInt(0) {
		return errs.NewError(errs.ErrNegativeLockTime, "negative sequence: %d", stackSequence.Int64())
	}

	sequence := stackSequence.Int64()

	// To provide for future soft-fork extensibility, if the
	// operand has the disabled lock-time flag set,
	// CHECKSEQUENCEVERIFY behaves as a NOP.
	if sequence&int64(bt.SequenceLockTimeDisabled) != 0 {
		return nil
	}

	// Transaction version numbers not high enough to trigger CSV rules must
	// fail.
	if t.tx.Version < 2 {
		return errs.NewError(errs.ErrUnsatisfiedLockTime, "invalid transaction version: %d", t.tx.Version)
	}

	// Sequence numbers with their most significant bit set are not
	// consensus constrained. Testing that the transaction's sequence
	// number does not have this bit set prevents using this property
	// to get around a CHECKSEQUENCEVERIFY check.
	txSequence := int64(t.tx.Inputs[t.inputIdx].SequenceNumber)
	if txSequence&int64(bt.SequenceLockTimeDisabled) != 0 {
		return errs.NewError(errs.ErrUnsatisfiedLockTime,
			"transaction sequence has sequence locktime disabled bit set: 0x%x", txSequence)
	}

	// Mask off non-consensus bits before doing comparisons.
	lockTimeMask := int64(bt.SequenceLockTimeIsSeconds | bt.SequenceLockTimeMask)

	return verifyLockTime(txSequence&lockTimeMask, bt.SequenceLockTimeIsSeconds, sequence&lockTimeMask)
}

// opcodeToAltStack removes the top item from the main data stack and pushes it
// onto the alternate data stack.
//
// Main data stack transformation: [... x1 x2 x3] -> [... x1 x2]
// Alt data stack transformation:  [... y1 y2 y3] -> [... y1 y2 y3 x3]
func opcodeToAltStack(op *ParsedOpcode, t *thread) error {
	so, err := t.dstack.PopByteArray()
	if err != nil {
		return err
	}

	t.astack.PushByteArray(so)

	return nil
}

// opcodeFromAltStack removes the top item from the alternate data stack and
// pushes it onto the main data stack.
//
// Main data stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 y3]
// Alt data stack transformation:  [... y1 y2 y3] -> [... y1 y2]
func opcodeFromAltStack(op *ParsedOpcode, t *thread) error {
	so, err := t.astack.PopByteArray()
	if err != nil {
		return err
	}

	t.dstack.PushByteArray(so)

	return nil
}

// opcode2Drop removes the top 2 items from the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1]
func opcode2Drop(op *ParsedOpcode, t *thread) error {
	return t.dstack.DropN(2)
}

// opcode2Dup duplicates the top 2 items on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x2 x3]
func opcode2Dup(op *ParsedOpcode, t *thread) error {
	return t.dstack.DupN(2)
}

// opcode3Dup duplicates the top 3 items on the data stack.
//
// Stack transformation: [... x1 x2 x3] -> [... x1 x2 x3 x1 x2 x3]
func opcode3Dup(op *ParsedOpcode, t *thread) error {
	return t.dstack.DupN(3)
}

// opcode2Over duplicates the 2 items before the top 2 items on the data stack.
//
// Stack transformation: [... x1 x2 x3 x4] -> [... x1 x2 