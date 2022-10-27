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
// The basic idea is, we have accounts, each with a master private key. If someone would like to