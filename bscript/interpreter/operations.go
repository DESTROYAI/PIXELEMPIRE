package interpreter

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // OP_SHA1 support requires this
	"crypto/sha256"
	"hash"
	"math/big"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-