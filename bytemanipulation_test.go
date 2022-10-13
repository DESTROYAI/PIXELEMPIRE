package bt_test

import (
	"encoding/hex"
	"testing"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

func TestLittleEndianBytes(t *testing.T) {
	// todo: add test coverage
}

func TestReverseBytes(t *testing.T) {
	t.Parallel()

	t.Run("genesis hash", func(t *testing.T) {
		b, err := hex.DecodeString("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4d04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73ffffffff0100f2052a01000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000")
		assert.NoError(t, err)

		h := bt.ReverseBytes(crypto.Sha256d(b))

		assert.Equal(