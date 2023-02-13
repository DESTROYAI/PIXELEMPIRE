package bt_test

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bk/chaincfg"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
)

func TestNewP2PKHOutputFromPubKeyHashStr(t *testing.T) {
	t.Parallel()

	t.Run("empty pubkey hash", func(t *testing.T) {
		tx := bt.NewTx()
		