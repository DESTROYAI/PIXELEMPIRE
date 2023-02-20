
package unlocker_test

import (
	"context"
	"testing"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
	"github.com/libsv/go-bt/v2/unlocker"
	"github.com/stretchr/testify/assert"
)

func TestLocalUnlocker_UnlockAllInputs(t *testing.T) {
	t.Parallel()

	incompleteTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	tx, err := bt.NewTxFromString(incompleteTx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Add the UTXO amount and script.
	tx.InputIdx(0).PreviousTxSatoshis = 100000000
	tx.InputIdx(0).PreviousTxScript, err = bscript.NewFromHexString("76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac")
	assert.NoError(t, err)

	// Our private key
	var w *wif.WIF
	w, err = wif.DecodeWIF("cNGwGSc7KRrTmdLUZ54fiSXWbhLNDc2Eg5zNucgQxyQCzuQ5YRDq")
	assert.NoError(t, err)

	unlocker := unlocker.Getter{PrivateKey: w.PrivKey}
	err = tx.FillAllInputs(context.Background(), &unlocker)
	assert.NoError(t, err)

	expectedSignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	assert.Equal(t, expectedSignedTx, tx.String())
	assert.NotEqual(t, incompleteTx, tx.String())
}

func TestLocalUnlocker_ValidSignature(t *testing.T) {
	tests := map[string]struct {
		tx *bt.Tx
	}{
		"valid signature 1": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From("45be95d2f2c64e99518ffbbce03fb15a7758f20ee5eecf0df07938d977add71d", 0, "76a914c7c6987b6e2345a6b138e3384141520a0fbc18c588ac", 15564838601))

				script1, err := bscript.NewFromHexString("76a91442f9682260509ac80722b1963aec8a896593d16688ac")
				assert.NoError(t, err)

				assert.NoError(t, tx.AddP2PKHOutputFromScript(script1, 375041432))

				script2, err := bscript.NewFromHexString("76a914c36538e91213a8100dcb2aed456ade363de8483f88ac")
				assert.NoError(t, err)

				assert.NoError(t, tx.AddP2PKHOutputFromScript(script2, 15189796941))

				return tx
			}(),
		},
		"valid signature 2": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()

				assert.NoError(
					t,
					tx.From("64faeaa2e3cbadaf82d8fa8c7ded508cb043c5d101671f43c084be2ac6163148", 1, "76a914343cadc47d08a14ef773d70b3b2a90870b67b3ad88ac", 5000000000),
				)
				tx.Inputs[0].SequenceNumber = 0xfffffffe

				script1, err := bscript.NewFromHexString("76a9140108b364bbbddb222e2d0fac1ad4f6f86b10317688ac")
				assert.NoError(t, err)

				assert.NoError(t, tx.AddP2PKHOutputFromScript(script1, 2200000000))

				script2, err := bscript.NewFromHexString("76a9143ac52294c730e7a4e9671abe3e7093d8834126ed88ac")
				assert.NoError(t, err)

				assert.NoError(t, tx.AddP2PKHOutputFromScript(script2, 2799998870))
				return tx
			}(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := test.tx

			var w *wif.WIF
			w, err := wif.DecodeWIF("cNGwGSc7KRrTmdLUZ54fiSXWbhLNDc2Eg5zNucgQxyQCzuQ5YRDq")
			assert.NoError(t, err)

			unlocker := &unlocker.Simple{PrivateKey: w.PrivKey}
			uscript, err := unlocker.UnlockingScript(context.Background(), tx, bt.UnlockerParams{})
			assert.NoError(t, err)