package bt

import "github.com/pkg/errors"

// General errors.
var (
	ErrInvalidTxID       = errors.New("invalid TxID")
	ErrTxNil             = errors.New("tx is nil")
	ErrTxTooShort        = errors.New("too short to be a tx - even an empty tx has 10 bytes")
	ErrNLockTimeLength   = errors.New("nLockTime length must be 4 bytes long")
	ErrEmptyValues       = erro