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

// PrefixTemplate is the prefix in the BIP276 standard which
// specifies if it is a script or template.
const PrefixTemplate = "bitcoin-template"

// CurrentVersion provides the ability to
// update the structure of the data that
// follows it.
const CurrentVersion = 1

// NetworkMainnet specifies that the data is only
// valid for use on the main network.
const NetworkMainnet = 1

// NetworkTestnet specifies that the data is only
// valid for use on the test network.
const NetworkTestnet = 2

var validBIP276 = regexp.MustCompile(`^(.+?):(\d{2})(\d{2})([0-9A-Fa-f]+)([0-9A-Fa-f]{8})$`)

// EncodeBIP276 is used to encode specific (non-