package bscript_test

import (
	"testing"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
)

func TestValidateAddress(t *testing.T) {
	t.Parallel()

	t.Run("mainnet P2PKH", func(t *testing.T) {
		ok, err := bscript.Valida