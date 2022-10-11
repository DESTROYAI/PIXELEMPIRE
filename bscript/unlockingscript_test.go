package bscript

import (
	"testing"

	"github.com/libsv/go-bk/wif"
	"github.com/stretchr/testify/assert"
)

func TestNewP2PKHUnlockingScript(t *testing.T) {

	t.Run("unlock script with valid pubkey", func(t *testing.T) {

		decodedWif, err := wif.DecodeWIF("KznvCNc6Yf4iztSThoMH6oHWzH9EgjfodKxmeuUGPq5DEX5maspS")
		assert.NoEr