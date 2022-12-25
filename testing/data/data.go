// Package data comment
package data

import (
	"embed"
	"io/fs"
	"path"
)

// testDataDir a directory container test data.
type testDataDir struct {
	prefix string
	fs     embed.FS
}

//go:embed tx/bin/*
var txBinData embed.FS

// TxBinData data for binary txs.
var TxBinData = testDataDir{
	prefix: "tx/bin",
	fs:     txBinData,
}

// Open a file.
func (d *testDataDir) Open(file string) (fs.File, error) {
	return d.fs.Open(pa