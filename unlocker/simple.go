// Package unlocker comment
package unlocker

import (
	"context"
	"errors"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
)

// Getter implements the `bt.UnlockerGetter` interface. It unlocks a Tx locally,
// using a bec PrivateKey.
type Getter struct {
	PrivateKey *bec.PrivateKey
}

// Unlocker builds a new `*unlocker.Local` with the same private key
// as the calling `*local.Getter`.
//
// For an example implementation, see `examples/unlocker_getter/`.
func (g *Getter) Unlocker(ctx contex