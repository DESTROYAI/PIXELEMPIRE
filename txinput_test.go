package bt_test

import (
	"context"
	"encoding/hex"
	"errors"
	"math"
	"testing"

	. "github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
	"github.com/libsv/go-bt/v2/unlocker"
	"github.com/stretchr/testify/assert"
)

func TestAddInputFromTx(t *testing.T) {
	pubkey1, _ := hex.DecodeString("0280f642908697e8068c2e921bd998d6c2b90553064656f91b9cb9e98f443aac30")
	pubkey2, _ := hex.DecodeString("02434dc3db4281c0895d7a126bb266e7648caca7d0e2e487bc41f954722d4ee397")

	prvTx := bt.NewTx()
	err := prvTx.AddP2PKHOutputFromPubKeyBytes(pubkey1, uint64(100000))
	assert.NoError(t, err)
	err = prvTx.AddP2PKHOutputFromPubKeyBytes(pubkey1, uint64(100000))
	assert.NoError(t, err)
	err = prvTx.AddP2PKHOutputFromPubKeyBytes(pubkey2, uint64(100000))
	assert.NoError(t, err)

	newTx := bt.NewTx()
	err = newTx.AddP2PKHInputsFromTx(prvTx, pubkey1)
	assert.NoError(t, err)
	assert.Equal(t, newTx.InputCount(), 2) // only 2 utxos added
	assert.Equal(t, newTx.TotalInputSatoshis(), uint64(200000))
}

func TestTx_InputCount(t *testing.T) {
	t.Run("get input count", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000,
		)
		assert.NoError(t, err)
		assert.Equal(t, 1, tx.InputCount())
	})
}

func TestTx_From(t *testing.T) {
	t.Run("invalid locking script (hex decode failed)", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"0",
			4000000,
		)
		assert.Error(t, err)

		err = tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae4016",
			4000000,
		)
		assert.Error(t, err)
	})

	t.Run("valid script and tx", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		err := tx.From(
			"07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b",
			0,
			"76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac",
			4000000,
		)
		assert.NoError(t, err)

		inputs := tx.Inputs
		assert.Equal(t, 1, len(inputs))
		assert.Equal(t, "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b", hex.EncodeToString(inputs[0].PreviousTxID()))
		assert.Equal(t, uint32(0), inputs[0].PreviousTxOutIndex)
		assert.Equal(t, uint64(4000000), inputs[0].PreviousTxSatoshis)
		assert.Equal(t, bt.DefaultSequenceNumber, inputs[0].SequenceNumber)
		assert.Equal(t, "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac", inputs[0].PreviousTxScript.String())
	})
}

func TestTx_FromUTXOs(t *testing.T) {
	t.Parallel()

	t.Run("one utxo", func(t *testing.T) {
		tx := bt.NewTx()
		script, err := bscript.NewFromHexString("76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac")
		assert.NoError(t, err)

		txID, err := hex.DecodeString("07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b")
		assert.NoError(t, err)

		assert.NoError(t, tx.FromUTXOs(&bt.UTXO{
			TxID:          txID,
			LockingScript: script,
			Vout:          0,
			Satoshis:      1000,
		}))

		input := tx.Inputs[0]
		assert.Equal(t, len(tx.Inputs), 1)
		assert.Equal(t, "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b", input.PreviousTxIDStr())
		assert.Equal(t, uint32(0), input.PreviousTxOutIndex)
		assert.Equal(t, uint64(1000), input.PreviousTxSatoshis)
		assert.Equal(t, "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac", input.PreviousTxScript.String())
	})

	t.Run("multiple utxos", func(t *testing.T) {
		tx := bt.NewTx()
		script, err := bscript.NewFromHexString("76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac")
		assert.NoError(t, err)
		txID, err := hex.DecodeString("07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b")
		assert.NoError(t, err)

		script2, err := bscript.NewFromHexString("76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac")
		assert.NoError(t, err)
		txID2, err := hex.DecodeString("3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5")
		assert.NoError(t, err)

		assert.NoError(t, tx.FromUTXOs(&bt.UTXO{
			TxID:          txID,
			LockingScript: script,
			Vout:          0,
			Satoshis:      1000,
		}, &bt.UTXO{
			TxID:          txID2,
			LockingScript: script2,
			Vout:          1,
			Satoshis:      2000,
		}))

		assert.Equal(t, len(tx.Inputs), 2)

		input := tx.Inputs[0]
		assert.Equal(t, "07912972e42095fe58daaf09161c5a5da57be47c2054dc2aaa52b30fefa1940b", input.PreviousTxIDStr())
		assert.Equal(t, uint32(0), input.PreviousTxOutIndex)
		assert.Equal(t, uint64(1000), input.PreviousTxSatoshis)
		assert.Equal(t, "76a914af2590a45ae401651fdbdf59a76ad43d1862534088ac", input.PreviousTxScript.String())

		input = tx.Inputs[1]
		assert.Equal(t, "3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5", input.PreviousTxIDStr())
		assert.Equal(t, uint32(1), input.PreviousTxOutIndex)
		assert.Equal(t, uint64(2000), input.PreviousTxSatoshis)
		assert.Equal(t, "76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac", input.PreviousTxScript.String())
	})
}

func TestTx_Fund(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tx                      *bt