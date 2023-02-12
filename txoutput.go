
package bt

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
)

// newOutputFromBytes returns a transaction Output from the bytes provided
func newOutputFromBytes(bytes []byte) (*Output, int, error) {
	if len(bytes) < 8 {
		return nil, 0, fmt.Errorf("%w < 8", ErrOutputTooShort)
	}

	offset := 8
	l, size := NewVarIntFromBytes(bytes[offset:])
	offset += size

	totalLength := offset + int(l)

	if len(bytes) < totalLength {
		return nil, 0, fmt.Errorf("%w < 8 + script", ErrInputTooShort)
	}

	s := bscript.Script(bytes[offset:totalLength])

	return &Output{
		Satoshis:      binary.LittleEndian.Uint64(bytes[0:8]),
		LockingScript: &s,
	}, totalLength, nil
}

// TotalOutputSatoshis returns the total Satoshis outputted from the transaction.
func (tx *Tx) TotalOutputSatoshis() (total uint64) {
	for _, o := range tx.Outputs {
		total += o.Satoshis
	}
	return
}

// AddP2PKHOutputFromPubKeyHashStr makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromPubKeyHashStr(publicKeyHash string, satoshis uint64) error {
	s, err := bscript.NewP2PKHFromPubKeyHashStr(publicKeyHash)
	if err != nil {
		return err
	}
