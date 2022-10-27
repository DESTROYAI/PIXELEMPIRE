package main

import (
	"context"

	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/unlocker"
)

// This example gives a simple in-memory based example of how to implement and use a `bt.UnlockerGetter`
// using derivated public/private keys.
//
// The basic idea is, we have accounts, each with a master private key. If someone would like to send money
// to an account, they request "destinations" from the account. These destinations are added to the
// tx, which is then ultimately broadcast.
//
// A destination in this example is simply a P2PKH locking script, however, the PK Hash will be unique
// on each call as under the hood, the account is deriving a new private/public key pair from its master
// key, creating a P2PKH script from that this pair, and storing the value used to derive this private/public
// key pair against the P2PKH script that it produced.
//
// When an account wishes 