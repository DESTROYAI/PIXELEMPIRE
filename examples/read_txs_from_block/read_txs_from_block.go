package main

import (
	"bufio"
	"fmt"
	"io"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/testing/data"
)

// In this example, all txs from a block are being read in via chunking, so at no point
// does the entire block have to be held in memory, and instead can be streamed.
//
// We represent the block by interactively readin