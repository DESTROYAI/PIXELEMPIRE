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
func (g *Getter) Unlocker(ctx context.Context, lockingScript *bscript.Script) (bt.Unlocker, error) {
	return &Simple{PrivateKey: g.PrivateKey}, nil
}

// Simple implements the a simple `bt.Unlocker` interface. It is used to build an unlocking script
// using a bec Private Key.
type Simple struct {
	PrivateKey *bec.PrivateKey
}

// UnlockingScript create the unlocking script for a given input using the PrivateKey passed in through the
// the `unlock.Local` struct.
//
// UnlockingScript generates and uses an ECDSA signature for the provided hash digest using the private key
// as well 