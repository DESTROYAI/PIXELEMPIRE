
package bscript

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"
	"strings"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bk/crypto"
)

// ScriptKey types.
const (
	ScriptTypePubKey      = "pubkey"
	ScriptTypePubKeyHash  = "pubkeyhash"
	ScriptTypeNonStandard = "nonstandard"
	ScriptTypeEmpty       = "empty"
	ScriptTypeSecureHash  = "securehash"
	ScriptTypeMultiSig    = "multisig"
	ScriptTypeNullData    = "nulldata"
)

// Script type
type Script []byte

// NewFromHexString creates a new script from a hex encoded string.
func NewFromHexString(s string) (*Script, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return NewFromBytes(b), nil
}

// NewFromBytes wraps a byte slice with the Script type.
func NewFromBytes(b []byte) *Script {
	s := Script(b)
	return &s
}

// NewFromASM creates a new script from a BitCoin ASM formatted string.
func NewFromASM(str string) (*Script, error) {
	s := Script{}

	for _, section := range strings.Split(str, " ") {
		if val, ok := opCodeStrings[section]; ok {
			_ = s.AppendOpcodes(val)
		} else {
			if err := s.AppendPushDataHexString(section); err != nil {
				return nil, ErrInvalidOpCode
			}
		}
	}

	return &s, nil
}

// NewP2PKHFromPubKeyEC takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyEC(pubKey *bec.PublicKey) (*Script, error) {
	return NewP2PKHFromPubKeyBytes(pubKey.SerialiseCompressed())
}

// NewP2PKHFromPubKeyStr takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyStr(pubKey string) (*Script, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	return NewP2PKHFromPubKeyBytes(pubKeyBytes)
}

// NewP2PKHFromPubKeyBytes takes public key bytes (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyBytes(pubKeyBytes []byte) (*Script, error) {
	if len(pubKeyBytes) != 33 {
		return nil, ErrInvalidPKLen
	}
	return NewP2PKHFromPubKeyHash(crypto.Hash160(pubKeyBytes))
}

// NewP2PKHFromPubKeyHash takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyHash(pubKeyHash []byte) (*Script, error) {
	b := []byte{
		OpDUP,
		OpHASH160,
		OpDATA20,
	}
	b = append(b, pubKeyHash...)
	b = append(b, OpEQUALVERIFY)
	b = append(b, OpCHECKSIG)

	s := Script(b)
	return &s, nil
}

// NewP2PKHFromPubKeyHashStr takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyHashStr(pubKeyHash string) (*Script, error) {
	hash, err := hex.DecodeString(pubKeyHash)
	if err != nil {
		return nil, err
	}

	return NewP2PKHFromPubKeyHash(hash)
}

// NewP2PKHFromAddress takes an address
// and creates a P2PKH script from it.
func NewP2PKHFromAddress(addr string) (*Script, error) {
	a, err := NewAddressFromString(addr)
	if err != nil {
		return nil, err
	}

	var publicKeyHashBytes []byte
	if publicKeyHashBytes, err = hex.DecodeString(a.PublicKeyHash); err != nil {
		return nil, err
	}

	s := new(Script)
	_ = s.AppendOpcodes(OpDUP, OpHASH160)
	if err = s.AppendPushData(publicKeyHashBytes); err != nil {
		return nil, err
	}
	_ = s.AppendOpcodes(OpEQUALVERIFY, OpCHECKSIG)

	return s, nil
}

// NewP2PKHFromBip32ExtKey takes a *bip32.ExtendedKey and creates a P2PKH script from it,
// using an internally random generated seed, returning the script and derivation path used.
func NewP2PKHFromBip32ExtKey(privKey *bip32.ExtendedKey) (*Script, string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, "", err
	}

	derivationPath := bip32.DerivePath(binary.LittleEndian.Uint64(b[:]))
	pubKey, err := privKey.DerivePublicKeyFromPath(derivationPath)
	if err != nil {
		return nil, "", err
	}

	lockingScript, err := NewP2PKHFromPubKeyBytes(pubKey)
	if err != nil {
		return nil, "", err
	}

	return lockingScript, derivationPath, nil
}

// AppendPushData takes data bytes and appends them to the script
// with proper PUSHDATA prefixes
func (s *Script) AppendPushData(d []byte) error {
	p, err := EncodeParts([][]byte{d})
	if err != nil {
		return err
	}

	*s = append(*s, p...)
	return nil
}

// AppendPushDataHexString takes a hex string and appends them to the
// script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataHexString(str string) error {
	h, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	return s.AppendPushData(h)
}

// AppendPushDataString takes a string and appends its UTF-8 encoding
// to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataString(str string) error {
	return s.AppendPushData([]byte(str))
}

// AppendPushDataArray takes an array of data bytes and appends them
// to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataArray(d [][]byte) error {
	p, err := EncodeParts(d)
	if err != nil {
		return err
	}

	*s = append(*s, p...)
	return nil
}

// AppendPushDataStrings takes an array of strings and appends their
// UTF-8 encoding to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataStrings(pushDataStrings []string) error {
	dataBytes := make([][]byte, 0)
	for _, str := range pushDataStrings {
		strBytes := []byte(str)
		dataBytes = append(dataBytes, strBytes)
	}