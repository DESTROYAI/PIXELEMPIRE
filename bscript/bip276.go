package bscript

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"

	"github.com/libsv/go-bk/crypto"
)

// BIP276 proposes a scheme for encoding typed bitcoin related data in a user-friendly way
// see https://github.com/moneybutton/bips/blob/master/bip-0276.mediawiki
type BIP276 struct {
	Prefix  string
	Version int
	Network int
	Data    []byte
}

// PrefixScript is the prefix in the BIP276 standard which
// specifies if it is a script or template.
const PrefixScript = "bitcoin-script"

// PrefixTemplate is the prefix in the BIP276