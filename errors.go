package bt

import "github.com/pkg/errors"

// General errors.
var (
	ErrInvalidTxID       = errors.New("invalid TxID")
	ErrTxNil             = errors.New("tx is nil")
	ErrTxTooShort        = errors.New("too short to be a tx - even an empty tx has 10 bytes")
	ErrNLockTimeLength   = errors.New("nLockTime length must be 4 bytes long")
	ErrEmptyValues       = errors.New("empty value or values passed, all arguments are required and cannot be empty")
	ErrUnsupportedScript = errors.New("non-P2PKH input used in the tx - unsupported")
	ErrInvalidScriptType = errors.New("invalid script type")
	ErrNoUnlocker        = errors.New("unlocker not supplied")
)

// Sentinal errors reported by inputs.
var (
	ErrInputNoExist  = errors.New("specified input does not exist")
	ErrInputTooShort = errors.New("input length too short")
)

// Sentinal errors reported by outputs.
var (
	ErrOutputNoExist  = errors.New("specified output does not exist")
	ErrOutputTooShort = errors.New("output length too short")
)

// Sentinal errors reported by change.
var (
	ErrInsufficientInputs = errors.New("satoshis inputted to the tx are less than the outputted satoshis")
)

// Sentinal errors reported by signature hash.
var (
	ErrEmptyPreviousTxID     = errors.New("'PreviousTxID' not supplied")
	ErrEmptyPreviousTxScript = errors.New("'PreviousTxScript' not supplied")
)

// Sentinel errors reported by the fees.
var (
	ErrFeeQuotesNotInit = errors.New("feeQuotes have not been setup, call NewFeeQuotes")
	ErrMinerNoQuotes    = errors.New("miner has no qu