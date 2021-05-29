package interpreter

import "math"

type config interface {
	AfterGenesis() bool
	MaxOps() int
	MaxStackSize() int
	MaxScriptSize() int
	MaxScriptE