package bt

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
)

/*
General format (inside a block) of each output of a transaction - Txout
Field	                        Description	                                Size
-----------------------------------------------------------------------------------------------------
value                         non-negative integer giving the number of   8 bytes
                              Satoshis(BTC/10^8) to be transferred
Txout-script length           non-negative integer                        1 - 9 bytes VI = VarInt
Txout-script / scriptPub