package interpreter

import "math"

type config interface {
	AfterGenesis() bool
	MaxOps() int
	MaxStackSize() int
	MaxScriptSize() int
	MaxScriptElementSize() int
	MaxScriptNumberLength() int
	MaxPubKeysPerMultiSig() int
}

// Limits applied to transactions before genesis
const (
	MaxOpsBeforeGenesis                = 500
	MaxStackSizeBeforeGenesis          = 1000
	MaxScriptSizeBeforeGenesis         = 10000
	MaxScriptElementSizeBeforeGenesis  = 520
	MaxScriptNumberLengthBeforeGenesis = 4
	MaxPubKeysPerMultiSigBeforeGenesis = 20
)

type beforeGenesisConfig struct{}
type afterGenesisConfig struct{}

func (a *afterGenesisConfig) AfterGenesis() bool {
	return true
}

func (b *beforeGenesisConfig) AfterGenesis() bool {
	return fals