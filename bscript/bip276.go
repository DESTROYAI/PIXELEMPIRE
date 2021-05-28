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

// EncodeBIP276 is used to encode specific (non-standard) scripts in BIP276 format.
// See https://github.com/moneybutton/bips/blob/master/bip-0276.mediawiki
func EncodeBIP276(script BIP276) string {
	if script.Version == 0 || script.Version > 255 || script.Network == 0 || script.Network > 255 {
		return "ERROR"
	}

	p, c := createBIP276(script)

	return p + c
}

func createBIP276(script BIP276) (string, string) {
	payload := fmt.Sprintf("%s:%.2x%.2x%x", script.Prefix, script.Network, script.Version, script.Data)
	return payload, hex.EncodeToString(crypto.Sha256d([]byte(payload))[:4])
}

// DecodeBIP276 is used to decode B