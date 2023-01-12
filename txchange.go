package bt

import (
	"github.com/libsv/go-bt/v2/bscript"
)

const (
	// DustLimit is the current minimum txo output accepted by miners.
	DustLimit = 1
)

// ChangeToAddress calculates the amount of fees needed to cover the transaction
// and adds the leftover change in a new P2PKH output using the address provided.
func (tx *Tx) ChangeToAddress(addr string, f *FeeQuote) error {
	s, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return err
	}

	return tx.Change(s, f)
}

// Change calculates the amount of fees needed to cover the transaction
//  and adds the leftover change in a new output using the script provided.
func (tx *Tx) Change(s *bscript.Script, f *FeeQuote) error {
	if _, _, err := tx.change(f, &changeOutput{
		lockingScript: s,
		newOutput:     true,
	}); err != nil {
		return err
	}
	return nil
}

// ChangeToExistingOutput will calculate fees and add them to a