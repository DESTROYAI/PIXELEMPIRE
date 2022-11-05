package bt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// FeeType is used to specify which
// type of fee is used depending on
// the type of tx data (eg: standard
// bytes or data bytes).
type FeeType string

const (
	// FeeTypeStandard is the fee type for standard tx parts
	FeeTypeStandard FeeType = "standard"

	// FeeTypeData is the fee type for data tx parts
	FeeTypeData FeeType = "data"
)

// FeeQuotes contains a list of miners and the current fees for each miner as well as their expiry.
//
// This can be used when getting fees from multiple miners, and you want to use the cheapest for example.
//
// Usage setup should be calling NewFeeQuotes(minerName).
type FeeQuotes struct {
	mu     sync.RWMutex
	quotes map[string]*FeeQuote
}

// NewFeeQuotes will set u