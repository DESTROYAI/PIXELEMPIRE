
package bt_test

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"reflect"
	"testing"

	"github.com/libsv/go-bk/wif"
	. "github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/testing/data"
	"github.com/libsv/go-bt/v2/unlocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTx(t *testing.T) {
	t.Parallel()

	t.Run("new tx, defaults", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		assert.IsType(t, &bt.Tx{}, tx)
		assert.Equal(t, uint32(1), tx.Version)
		assert.Equal(t, uint32(0), tx.LockTime)
		assert.Equal(t, 0, tx.InputCount())
		assert.Equal(t, 0, tx.OutputCount())
		assert.Equal(t, uint64(0), tx.TotalOutputSatoshis())
		assert.Equal(t, uint64(0), tx.TotalInputSatoshis())
	})
}

func TestNewTxFromString(t *testing.T) {
	t.Parallel()

	t.Run("valid tx no Inputs", func(t *testing.T) {
		tx, err := bt.NewTxFromString("01000000000100000000000000001a006a07707265666978310c6578616d706c65206461746102133700000000")
		assert.NoError(t, err)
		assert.NotNil(t, tx)
	})

	t.Run("invalid tx", func(t *testing.T) {
		tx, err := bt.NewTxFromString("0")
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("invalid tx - too short", func(t *testing.T) {
		tx, err := bt.NewTxFromString("000000")
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("valid tx, 1 input, 1 output", func(t *testing.T) {
		rawTx := "02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c000000004847304402205cc711985ce2a6d61eece4f9b6edd6337bad3b7eca3aa3ce59bc15620d8de2a80220410c92c48a226ba7d5a9a01105524097f673f31320d46c3b61d2378e6f05320041ffffffff01c0aff629010000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000"
		tx, err := bt.NewTxFromString(rawTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Check version, locktime, Inputs
		assert.Equal(t, uint32(2), tx.Version)
		assert.Equal(t, uint32(0), tx.LockTime)
		assert.Equal(t, 1, len(tx.Inputs))

		// Create a new unlocking script
		//ptid, _ := hex.DecodeString("9c5b1428aaad5e9b0196c89be8628b366f33c7b22933da0489b921d487a7cb1c")
		i := &bt.Input{
			PreviousTxOutIndex: 0,
			SequenceNumber:     bt.DefaultSequenceNumber,
		}
		assert.NoError(t, i.PreviousTxIDAdd(tx.InputIdx(0).PreviousTxID()))
		i.UnlockingScript, err = bscript.NewFromHexString("47304402205cc711985ce2a6d61eece4f9b6edd6337bad3b7eca3aa3ce59bc15620d8de2a80220410c92c48a226ba7d5a9a01105524097f673f31320d46c3b61d2378e6f05320041")
		assert.NoError(t, err)
		assert.NotNil(t, i.UnlockingScript)

		// Check input type
		assert.Equal(t, tx.InputIdx(0), i)

		// Check output
		assert.Equal(t, 1, len(tx.Outputs))

		// New output
		var lscript *bscript.Script
		lscript, err = bscript.NewFromHexString("76a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac")
		assert.NoError(t, err)
		assert.NotNil(t, lscript)

		// Check the type
		o := bt.Output{Satoshis: 4999000000, LockingScript: lscript}
		assert.Equal(t, true, reflect.DeepEqual(*tx.Outputs[0], o))
	})
}

func TestNewTxFromBytes(t *testing.T) {
	t.Parallel()

	t.Run("valid tx", func(t *testing.T) {
		rawTx := "02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c0000000049483045022100c4df63202a9aa2bea5c24ebf4418d145e81712072ef744a4b108174f1ef59218022006eb54cf904707b51625f521f8ed2226f7d34b62492ebe4ddcb1c639caf16c3c41ffffffff0140420f00000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000"
		b, err := hex.DecodeString(rawTx)
		assert.NoError(t, err)

		var tx *bt.Tx
		tx, err = bt.NewTxFromBytes(b)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
	})

	t.Run("invalid tx, too short", func(t *testing.T) {
		rawTx := "000000"
		b, err := hex.DecodeString(rawTx)
		assert.NoError(t, err)

		var tx *bt.Tx
		tx, err = bt.NewTxFromBytes(b)
		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}

func TestTx_TxID(t *testing.T) {
	t.Parallel()

	t.Run("valid tx id", func(t *testing.T) {
		tx, err := bt.NewTxFromString("010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000")
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, "19dcf16ecc9286c3734fdae3d45d4fc4eb6b25f841131e06460f4939bba0026e", tx.TxID())
	})

	t.Run("new tx, no data, but has default tx id", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		assert.Equal(t, "d21633ba23f70118185227be58a63527675641ad37967e2aa461559f577aec43", tx.TxID())
	})
}

func TestVersion(t *testing.T) {
	t.Parallel()

	rawTx := "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	tx, err := bt.NewTxFromString(rawTx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, uint32(1), tx.Version)
}

func TestTx_IsCoinbase(t *testing.T) {
	t.Parallel()

	t.Run("invalid number of Inputs", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)
		assert.Equal(t, false, tx.IsCoinbase())
	})

	t.Run("valid coinbase tx, 1 input", func(t *testing.T) {
		rawTx := "02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0e5101010a2f4542323030302e302fffffffff0100f2052a01000000232103db233bb9fc387d78b133ec904069d46e95ff17da657671b44afa0bc64e89ac18ac00000000"
		tx, err := bt.NewTxFromString(rawTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		assert.Equal(t, true, tx.IsCoinbase())
		assert.Equal(t, 1, tx.InputCount())
	})

	t.Run("valid coinbase tx", func(t *testing.T) {
		coinbaseTx := "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4303bfea07322f53696d6f6e204f726469736820616e642053747561727420467265656d616e206d61646520746869732068617070656e2f9a46434790f7dbdea3430000ffffffff018a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"
		tx, err := bt.NewTxFromString(coinbaseTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, true, tx.IsCoinbase())
	})

	t.Run("tx is not a coinbase tx", func(t *testing.T) {
		coinbaseTx := "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
		tx, err := bt.NewTxFromString(coinbaseTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, false, tx.IsCoinbase())
	})

	t.Run("tx (2) is not a coinbase tx", func(t *testing.T) {
		coinbaseTx := "010000000159ef0cbb7881f2c934d6fb669f68f7c6a9c632f997152f828d1153806b7ac82b010000006b483045022100e775a21994cc6d6d6bf79d295aeea592e7b4cf8d8ecddaf67bb6626d7af82fd302201921a313de67e23a78c81dd5fe9a19322839c0ea1034b9c54e8206dea3aa9e68412103d1c02ee3522ff58df6c6287e67202a797b562fa8b5a9ed86613fe5ee48fb8821ffffffff02000000000000000011006a0e6d657461737472656d652e636f6dc9990200000000001976a914fa1b02ff7e41975d698fec6fb1b2d7e4656f8e7f88ac00000000"
		tx, err := bt.NewTxFromString(coinbaseTx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, false, tx.IsCoinbase())
	})
}

func TestTx_CreateTx(t *testing.T) {
	t.Parallel()

	tx := bt.NewTx()
	assert.NotNil(t, tx)

	err := tx.From(
		"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
		0,
		"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
		2000000,
	)
	assert.NoError(t, err)

	err = tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1999942)
	assert.NoError(t, err)

	var wif *WIF
	wif, err = DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
	assert.NoError(t, err)
	assert.NotNil(t, wif)

	err = tx.FillAllInputs(context.Background(), &unlocker.Getter{PrivateKey: wif.PrivKey})
	assert.NoError(t, err)
}

func TestTx_HasDataOutputs(t *testing.T) {
	t.Parallel()

	t.Run("has data Outputs", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			0,
			"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
			2000000,
		)
		assert.NoError(t, err)

		err = tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1999942)
		assert.NoError(t, err)

		// Add op return data
		type OpReturnData [][]byte
		ops := OpReturnData{[]byte("prefix1"), []byte("example data"), []byte{0x13, 0x37}}

		err = tx.AddOpReturnPartsOutput(ops)
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		err = tx.FillAllInputs(context.Background(), &unlocker.Getter{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, true, tx.HasDataOutputs())
	})

	t.Run("no data Outputs", func(t *testing.T) {
		tx := bt.NewTx()
		assert.NotNil(t, tx)

		err := tx.From(
			"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
			0,
			"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
			2000000,
		)
		assert.NoError(t, err)

		err = tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mk", 1999942)
		assert.NoError(t, err)

		var wif *WIF
		wif, err = DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
		assert.NoError(t, err)
		assert.NotNil(t, wif)

		err = tx.FillAllInputs(context.Background(), &unlocker.Getter{PrivateKey: wif.PrivKey})
		assert.NoError(t, err)

		assert.Equal(t, false, tx.HasDataOutputs())
	})
}

func TestTx_OutputIdx(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tx        *bt.Tx
		idx       int
		expOutput *bt.Output
	}{
		"tx with 3 Outputs and output idx 0 requested should return output": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.PayToAddress("myUmQeCYxQECGHXbupe539n41u6BTBz1Eh", 1000))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mz", 1000))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mz", 1000))
				return tx
			}(),
			idx: 0,
			expOutput: &bt.Output{
				Satoshis: 1000,
				LockingScript: func() *bscript.Script {
					s, err := bscript.NewP2PKHFromAddress("myUmQeCYxQECGHXbupe539n41u6BTBz1Eh")
					assert.NoError(t, err)
					return s
				}(),
			},
		}, "tx with 3 Outputs and output idx 2 requested should return output": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.PayToAddress("myUmQeCYxQECGHXbupe539n41u6BTBz1Eh", 1000))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mz", 1000))
				assert.NoError(t, tx.PayToAddress("mywmGVP89x3DsLNqk3NvctfQy9m9pvt7mz", 1000))
				return tx
			}(),
			idx: 2,
			expOutput: &bt.Output{
				Satoshis: 1000,
				LockingScript: func() *bscript.Script {
					s, err := bscript.NewP2PKHFromAddress("mywmGVP89x3DsLNqk3NvctfQy9m9pvt7mz")
					assert.NoError(t, err)
					return s
				}(),
			},
		}, "tx with 3 Outputs and output idx 5 requested should return nil": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.PayToAddress("myUmQeCYxQECGHXbupe539n41u6BTBz1Eh", 1000))
				assert.NoError(t, tx.PayToAddress("n2wmGVP89x3DsLNqk3NvctfQy9m9pvt7mz", 1000))
				assert.NoError(t, tx.PayToAddress("mywmGVP89x3DsLNqk3NvctfQy9m9pvt7mz", 1000))
				return tx
			}(),
			idx:       5,
			expOutput: nil,
		}, "tx with 0 Outputs and output idx 5 requested should return nil": {
			tx: func() *bt.Tx {
				return bt.NewTx()
			}(),
			idx:       5,
			expOutput: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			o := test.tx.OutputIdx(test.idx)
			assert.Equal(t, test.expOutput, o)
		})
	}
}

func TestTx_InputIdx(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tx       *bt.Tx
		idx      int
		expInput *bt.Input
	}{
		"tx with 3 Inputs and input idx 0 requested should return correct input": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					1000,
				))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000,
				))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000,
				))
				return tx
			}(),
			idx: 0,
			expInput: func() *bt.Input {
				in := &bt.Input{
					PreviousTxSatoshis: 1000,
					PreviousTxScript: func() *bscript.Script {
						b, err := bscript.NewFromHexString("76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac")
						assert.NoError(t, err)
						return b
					}(),
					PreviousTxOutIndex: 0,
					SequenceNumber:     bt.DefaultSequenceNumber,
				}
				_ = in.PreviousTxIDAddStr("3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5")
				return in
			}(),
		}, "tx with 3 Outputs and output idx 2 requested should return output": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					1000,
				))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdad5",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					2000000,
				))
				assert.NoError(t, tx.From(
					"3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdac4",
					0,
					"76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac",
					999,
				))
				return tx
			}(),
			idx: 2,
			expInput: func() *bt.Input {
				in := &bt.Input{
					PreviousTxSatoshis: 999,
					PreviousTxScript: func() *bscript.Script {
						b, err := bscript.NewFromHexString("76a914eb0bd5edba389198e73f8efabddfc61666969ff788ac")
						assert.NoError(t, err)
						return b
					}(),
					PreviousTxOutIndex: 0,
					SequenceNumber:     bt.DefaultSequenceNumber,
				}
				_ = in.PreviousTxIDAddStr("3c8edde27cb9a9132c22038dac4391496be9db16fd21351565cc1006966fdac4")
				return in
			}(),
		}, "tx with 3 Outputs and output idx 5 requested should return nil": {
			tx: func() *bt.Tx {