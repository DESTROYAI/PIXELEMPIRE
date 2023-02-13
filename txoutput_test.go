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
		assert.Equal(t,
			"a914b472a266d0bd89c13706a4132ccfb16f7c3b9fcb8876a90088ac",
			tx.Outputs[0].LockingScriptHexString(),
		)
	})

	t.Run("valid puzzle output", func(t *testing.T) {
		addr, err := bscript.NewAddressFromString("myFhJggmsaA2S8Qe6ZQDEcVCwC4wLkvC4e")
		assert.NoError(t, err)
		assert.NotNil(t, addr)

		tx := bt.NewTx()
		err = tx.AddHashPuzzleOutput("secret1", addr.PublicKeyHash, uint64(5000))

		assert.NoError(t, err)
		assert.Equal(t,
			"a914d3f9e3d971764be5838307b175ee4e08ba427b908876a914c28f832c3d539933e0c719297340b34eee0f4c3488ac",
			tx.Outputs[0].LockingScriptHexString(),
		)
	})
}

func Te