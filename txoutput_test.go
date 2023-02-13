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
		err := tx.AddP2PKHOutputFromPubKeyHashStr(
			"",
			uint64(5000),
		)
		assert.NoError(t, err)
		assert.Equal(t,
			"76a91488ac",
			tx.Outputs[0].LockingScriptHexString(),
		)
	})

	t.Run("invalid pubkey hash", func(t *testing.T) {
		tx := bt.NewTx()
		err := tx.AddP2PKHOutputFromPubKeyHashStr(
			"0",
			uint64(5000),
		)
		assert.Error(t, err)
	})

	t.Run("valid output", func(t *testing.T) {
		// This is the PKH for address mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd
		tx := bt.NewTx()
		err := tx.AddP2PKHOutputFromPubKeyHashStr(
			"8fe80c75c9560e8b56ed64ea3c26e18d2c52211b",
			uint64(5000),
		)
		assert.NoError(t, err)
		assert.Equal(t,
			"76a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac",
			tx.Outputs[0].LockingScriptHexString(),
		)
	})
}

func TestNewHashPuzzleOutput(t *testing.T) {
	t.Parallel()

	t.Run("invalid public key", func(t *testing.T) {
		tx := bt.NewTx()
		err := tx.AddHashPuzzleOutput("", "0", uint64(5000))
		assert.Error(t, err)
	})

	t.Run("missing secret and public key", func(t *testing.T) {
		tx := bt.NewTx()
		err := tx.AddHashPuzzleOutput("", "", uint64(5000))

		assert.NoError(t, err)
		asser