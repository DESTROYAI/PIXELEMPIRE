
package bscript_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bk/chaincfg"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/libsv/go-bt/v2/bscript"
)

func TestNewP2PKHFromPubKeyStr(t *testing.T) {
	t.Parallel()

	scriptP2PKH, err := bscript.NewP2PKHFromPubKeyStr(
		"023717efaec6761e457f55c8417815505b695209d0bbfed8c3265be425b373c2d6",
	)
	assert.NoError(t, err)
	assert.NotNil(t, scriptP2PKH)
	assert.Equal(t,
		"76a9144d5d1920331b71735a97a606d9734aed83cb3dfa88ac",
		hex.EncodeToString(*scriptP2PKH),
	)
}

func TestNewP2PKHFromPubKey(t *testing.T) {
	t.Parallel()

	pk, _ := hex.DecodeString("023717efaec6761e457f55c8417815505b695209d0bbfed8c3265be425b373c2d6")

	pubkey, err := bec.ParsePubKey(pk, bec.S256())
	assert.NoError(t, err)

	scriptP2PKH, err := bscript.NewP2PKHFromPubKeyEC(pubkey)
	assert.NoError(t, err)
	assert.NotNil(t, scriptP2PKH)
	assert.Equal(t,
		"76a9144d5d1920331b71735a97a606d9734aed83cb3dfa88ac",
		hex.EncodeToString(*scriptP2PKH),
	)
}

func TestNewP2PKHFromBip32ExtKey(t *testing.T) {
	t.Parallel()

	t.Run("output is added", func(t *testing.T) {
		var b [64]byte
		_, err := rand.Read(b[:])
		assert.NoError(t, err)

		key, err := bip32.NewMaster(b[:], &chaincfg.TestNet)
		assert.NoError(t, err)

		script, derivationPath, err := bscript.NewP2PKHFromBip32ExtKey(key)

		assert.NoError(t, err)
		assert.NotEmpty(t, derivationPath)
		assert.NotNil(t, script)
		assert.True(t, script.IsP2PKH())
	})

	t.Run("invalid key errors", func(t *testing.T) {
		var b [64]byte
		_, err := rand.Read(b[:])
		assert.NoError(t, err)

		script, derivationPath, err := bscript.NewP2PKHFromBip32ExtKey(&bip32.ExtendedKey{})

		assert.Error(t, err)
		assert.Empty(t, derivationPath)
		assert.Nil(t, script)
	})
}

func TestNewFromHexString(t *testing.T) {
	t.Parallel()

	s, err := bscript.NewFromHexString("76a914e2a623699e81b291c0327f408fea765d534baa2a88ac")
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t,
		"76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
		hex.EncodeToString(*s),
	)
}

func TestScript_ToASM(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		script string
		expASM string
	}{
		"valid script": {
			script: "76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
			expASM: "OP_DUP OP_HASH160 e2a623699e81b291c0327f408fea765d534baa2a OP_EQUALVERIFY OP_CHECKSIG",
		},
		"empty script:": {
			script: "",
			expASM: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s, err := bscript.NewFromHexString(test.script)
			assert.NoError(t, err)

			asm, err := s.ToASM()
			assert.NoError(t, err)

			assert.Equal(t, test.expASM, asm)
		})
	}
}

func TestNewFromASM(t *testing.T) {
	t.Parallel()

	s, err := bscript.NewFromASM("OP_DUP OP_HASH160 e2a623699e81b291c0327f408fea765d534baa2a OP_EQUALVERIFY OP_CHECKSIG")
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t,
		"76a914e2a623699e81b291c0327f408fea765d534baa2a88ac",
		hex.EncodeToString(*s),
	)
}

func TestScript_IsP2PKH(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString("76a91403ececf2d12a7f614aef4c82ecf13c303bd9975d88ac")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsP2PKH())
}

func TestScript_IsP2PK(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString("2102f0d97c290e79bf2a8660c406aa56b6f189ff79f2245cc5aff82808b58131b4d5ac")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsP2PK())
}

func TestScript_IsP2SH(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString("a9149de5aeaff9c48431ba4dd6e8af73d51f38e451cb87")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsP2SH())
}

func TestScript_IsData(t *testing.T) {
	t.Parallel()

	b, err := hex.DecodeString("006a04ac1eed884d53027b2276657273696f6e223a22302e31222c22686569676874223a3634323436302c22707265764d696e65724964223a22303365393264336535633366376264393435646662663438653761393933393362316266623366313166333830616533306432383665376666326165633561323730222c22707265764d696e65724964536967223a2233303435303232313030643736333630653464323133333163613836663031386330343665353763393338663139373735303734373333333533363062653337303438636165316166333032323030626536363034353430323162663934363465393966356139353831613938633963663439353430373539386335396234373334623266646234383262663937222c226d696e65724964223a22303365393264336535633366376264393435646662663438653761393933393362316266623366313166333830616533306432383665376666326165633561323730222c2276637478223a7b2274784964223a2235373962343335393235613930656533396133376265336230306239303631653734633330633832343133663664306132303938653162656137613235313566222c22766f7574223a307d2c226d696e6572436f6e74616374223a7b22656d61696c223a22696e666f407461616c2e636f6d222c226e616d65223a225441414c20446973747269627574656420496e666f726d6174696f6e20546563686e6f6c6f67696573222c226d65726368616e74415049456e64506f696e74223a2268747470733a2f2f6d65726368616e746170692e7461616c2e636f6d2f227d7d46304402206fd1c6d6dd32cc85ddd2f30bc068445dd901c6bd85e394e45bb254716d2bb228022041f0f8b1b33c2e3702aee4ad47155548045ed945738b43dc0faed2e86faa12e4")
	assert.NoError(t, err)

	scriptPub := bscript.NewFromBytes(b)
	assert.NotNil(t, scriptPub)
	assert.Equal(t, true, scriptPub.IsData())
}

func TestScript_IsMultisigOut(t *testing.T) { // TODO: check this
	t.Parallel()

	t.Run("is multisig", func(t *testing.T) {
		b, err := hex.DecodeString("5201110122013353ae")
		assert.NoError(t, err)

		scriptPub := bscript.NewFromBytes(b)
		assert.NotNil(t, scriptPub)
		assert.Equal(t, true, scriptPub.IsMultiSigOut())
	})

	t.Run("is not multisig and no error", func(t *testing.T) {
		//Test Txid:de22e20422dbba8e8eeab87d5532480499abb01d6619bb66fe374f4d4a7500ee, vout:1

		b, err := hex.DecodeString("5101400176018801a901ac615e7961007901687f7700005279517f75007f77007901fd8763615379537f75517f77007901007e81517a7561537a75527a527a5379535479937f75537f77527a75517a67007901fe8763615379557f75517f77007901007e81517a7561537a75527a527a5379555479937f75557f77527a75517a67007901ff8763615379597f75517f77007901007e81517a7561537a75527a527a5379595479937f75597f77527a75517a67615379517f75007f77007901007e81517a7561537a75527a527a5379515479937f75517f77527a75517a6868685179517a75517a75517a75517a7561517a7561007982770079011494527951797f77537952797f750001127900a063610113795a7959797e01147e51797e5a797e58797e517a7561610079011479007958806152790079827700517902fd009f63615179515179517951938000795179827751947f75007f77517a75517a75517a7561517a75675179030000019f6301fd615279525179517951938000795179827751947f75007f77517a75517a75517a75617e517a756751790500000000019f6301fe615279545179517951938000795179827751947f75007f77517a75517a75517a75617e517a75675179090000000000000000019f6301ff615279585179517951938000795179827751947f75007f77517a75517a75517a75617e517a7568686868007953797e517a75517a75517a75617e517a75517a7561527951797e537a75527a527a527975757568607900a06351790112797e610079011279007958806152790079827700517902fd009f63615179515179517951938000795179827751947f75007f77517a75517a75517a7561517a75675179030000019f6301fd615279525179517951938000795179827751947f75007f77517a75517a75517a75617e517a756751790500000000019f6301fe615279545179517951938000795179827751947f75007f77517a75517a75517a75617e517a75675179090000000000000000019f6301ff615279585179517951938000795179827751947f75007f77517a75517a75517a75617e517a7568686868007953797e517a75517a75517a75617e517a75517a7561527951797e537a75527a527a5279757575685e7900a063615f795a7959797e01147e51797e5a797e58797e517a75616100796079007958806152790079827700517902fd009f63615179515179517951938000795179827751947f75007f77517a75517a75517a7561517a75675179030000019f6301fd615279525179517951938000795179827751947f75007f77517a75517a75517a75617e517a756751790500000000019f6301fe615279545179517951938000795179827751947f75007f77517a75517a75517a75617e517a75675179090000000000000000019f6301ff615279585179517951938000795179827751947f75007f77517a75517a75517a75617e517a7568686868007953797e517a75517a75517a75617e517a75517a7561527951797e537a75527a527a5279757575685c7900a063615d795a7959797e01147e51797e5a797e58797e517a75616100795e79007958806152790079827700517902fd009f63615179515179517951938000795179827751947f75007f77517a75517a75517a7561517a75675179030000019f6301fd615279525179517951938000795179827751947f75007f77517a75517a75517a75617e517a756751790500000000019f6301fe615279545179517951938000795179827751947f75007f77517a75517a75517a75617e517a75675179090000000000000000019f6301ff615279585179517951938000795179827751947f75007f77517a75517a75517a75617e517a7568686868007953797e517a75517a75517a75617e517a75517a7561527951797e537a75527a527a5279757575680079aa007961011679007982775179517958947f7551790128947f77517a75517a75618769011679a954798769011779011779ac69610115796100792097dfd76851bf465e8f715593b217714858bbe9570ff3bd5e33840a34e20ff0262102ba79df5f8ae7604a9830f03c7933028186aede0675a16f025dc4f8be8eec0382210ac407f0e4bd44bfc207355a778b046225a7068fc59ee7eda43ad905aadbffc800206c266b30e6a1319c66dc401e5bd6b432ba49688eecd118297041da8074ce0810201008ce7480da41702918d1ec8e6849ba32b4d65b1e40dc669c31a1e6306b266c011379011379855679aa616100790079517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e01007e81517a756157795679567956795679537956795479577995939521414136d08c5ed2bf3ba048afe6dcaebafeffffffffffffffffffffffffffffff0061517951795179517997527a75517a5179009f635179517993527a75517a685179517a75517a7561527a75517a517951795296a0630079527994527a75517a68537982775279827754527993517993013051797e527e53797e57797e527e52797e5579517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f517f7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7c7e7e56797e0079517a75517a75517a75517a75517a75517a75517a75517a75517a75517a75517a75517a756100795779ac517a75517a75517a75517a75517a75517a75517a75517a75517a7561517a75617777777777777777777777777777777777777777777777776ae0cfa0c0930b63270459fe368d5ed31da74c00de")
		assert.NoError(t, err)

		scriptPub := bscript.NewFromBytes(b)
		assert.NotNil(t, scriptPub)
		assert.Equal(t, false, scriptPub.IsMultiSigOut())
	})
}

func TestScript_PublicKeyHash(t *testing.T) {
	t.Parallel()

	t.Run("get as bytes", func(t *testing.T) {
		b, err := hex.DecodeString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
		assert.NoError(t, err)

		s := bscript.NewFromBytes(b)
		assert.NotNil(t, s)

		var pkh []byte
		pkh, err = s.PublicKeyHash()
		assert.NoError(t, err)
		assert.Equal(t, "04d03f746652cfcb6cb55119ab473a045137d265", hex.EncodeToString(pkh))
	})

	t.Run("get as string", func(t *testing.T) {
		s, err := bscript.NewFromHexString("76a91404d03f746652cfcb6cb55119ab473a045137d26588ac")
		assert.NoError(t, err)
		assert.NotNil(t, s)

		var pkh []byte
		pkh, err = s.PublicKeyHash()
		assert.NoError(t, err)
		assert.Equal(t, "04d03f746652cfcb6cb55119ab473a045137d265", hex.EncodeToString(pkh))
	})

	t.Run("empty script", func(t *testing.T) {
		s := &bscript.Script{}

		_, err := s.PublicKeyHash()
		assert.Error(t, err)
		assert.EqualError(t, err, "script is empty")
	})

	t.Run("nonstandard script", func(t *testing.T) {
		// example tx 37d4cc9f8a3b62e7f2e7c97c07a3282bfa924739c0e174733ff1b764ef8e3ebc
		s, err := bscript.NewFromHexString("76")
		assert.NoError(t, err)
		assert.NotNil(t, s)

		_, err = s.PublicKeyHash()
		assert.Error(t, err)
		assert.EqualError(t, err, "not a P2PKH")
	})
}

func TestErrorIsAppended(t *testing.T) {
	script, _ := hex.DecodeString("6a0548656c6c6f0548656c6c")
	s := bscript.Script(script)

	asm, err := s.ToASM()
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(asm, "[error]"), "toASM() should end with [error]")
}

func TestScript_AppendOpcodes(t *testing.T) {
	tests := map[string]struct {
		script    string
		appends   []byte
		expScript string
		expErr    error
	}{
		"successful single append": {
			script:    "OP_2 OP_2 OP_ADD",
			appends:   []byte{bscript.OpEQUALVERIFY},
			expScript: "OP_2 OP_2 OP_ADD OP_EQUALVERIFY",
		},
		"successful multiple append": {
			script:    "OP_2 OP_2 OP_ADD",
			appends:   []byte{bscript.OpEQUAL, bscript.OpVERIFY},
			expScript: "OP_2 OP_2 OP_ADD OP_EQUAL OP_VERIFY",
		},
		"unsuccessful push adata append": {
			script:  "OP_2 OP_2 OP_ADD",
			appends: []byte{bscript.OpEQUAL, bscript.OpPUSHDATA1, 0x44},
			expErr:  bscript.ErrInvalidOpcodeType,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			script, err := bscript.NewFromASM(test.script)
			assert.NoError(t, err)

			err = script.AppendOpcodes(test.appends...)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, test.expErr, errors.Unwrap(err).Error())
			} else {
				assert.NoError(t, err)
				asm, err := script.ToASM()
				assert.NoError(t, err)
				assert.Equal(t, test.expScript, asm)
			}
		})
	}
}

func TestScript_Equals(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		script1 *bscript.Script
		script2 *bscript.Script
		exp     bool
	}{
		"P2PKH scripts that equal should return true": {
			script1: func() *bscript.Script {
				s, err := bscript.NewP2PKHFromAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk")
				assert.NoError(t, err)
				return s
			}(),
			script2: func() *bscript.Script {
				s, err := bscript.NewP2PKHFromAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk")
				assert.NoError(t, err)
				return s
			}(),
			exp: true,
		}, "scripts from bytes equal should return true": {
			script1: func() *bscript.Script {
				b, err := hex.DecodeString("5201110122013353ae")
				assert.NoError(t, err)

				return bscript.NewFromBytes(b)
			}(),
			script2: func() *bscript.Script {
				b, err := hex.DecodeString("5201110122013353ae")
				assert.NoError(t, err)
