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
func (tx *Tx) Chang