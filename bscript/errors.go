package bscript

import "github.com/pkg/errors"

// Sentinel errors raised by data ops.
var (
	ErrDataTooBig   = errors.New("data too big")
	ErrDataTooSmall = errors.New("not enough data")
	ErrPartTooBig   = errors.New("part too big")
)

