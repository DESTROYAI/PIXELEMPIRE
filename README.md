# go-bt

> The go-to Bitcoin Transaction (BT) GoLang library  

[![Release](https://img.shields.io/github/release-pre/libsv/go-bt.svg?logo=github&style=flat&v=1)](https://github.com/libsv/go-bt/releases)
[![Build Status](https://img.shields.io/github/workflow/status/libsv/go-bt/run-go-tests?logo=github&v=3)](https://github.com/libsv/go-bt/actions)
[![Report](https://goreportcard.com/badge/github.com/libsv/go-bt?style=flat&v=1)](https://goreportcard.com/report/github.com/libsv/go-bt)
[![codecov](https://codecov.io/gh/libsv/go-bt/branch/master/graph/badge.svg?v=1)](https://codecov.io/gh/libsv/go-bt)
[![Go](https://img.shields.io/github/go-mod/go-version/libsv/go-bt?v=1)](https://golang.org/)
[![Sponsor](https://img.shields.io/badge/sponsor-libsv-181717.svg?logo=github&style=flat&v=3)](https://github.com/sponsors/libsv)
[![Donate](https://img.shields.io/badge/donate-bitcoin-ff9900.svg?logo=bitcoin&style=flat&v=3)](https://gobitcoinsv.com/#sponsor)
[![Mergify Status][mergify-status]][mergify]

[mergify]: https://mergify.io
[mergify-status]: https://img.shields.io/endpoint.svg?url=https://gh.mergify.io/badges/libsv/go-bt&style=flat
<br/>

## Table of Contents

- [Installation](#installation)
- [Documentation](#documentation)
- [Examples & Tests](#examples--tests)
- [Benchmarks](#benchmarks)
- [Code Standards](#code-standards)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

<br/>

## Installation

**go-bt** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).

```shell script
go get -u github.com/libsv/go-bt
```

<br/>

## Documentation

View the generated [documentation](https://pkg.go.dev/github.com/libsv/go-bt)

[![GoDoc](https://godoc.org/github.com/libsv/go-bt?status.svg&style=flat)](https://pkg.go.dev/github.com/libsv/go-bt)

For more information around the technical aspects of Bitcoin, please see the updated [Bitcoin Wiki](https://wiki.bitcoinsv.io/index.php/Main_Page)

<br/>

### Features

- Full featured Bitcoin transactions and transaction manipulation/functionality
- Auto-fee calculations for change outputs
- Transaction fee calculation and related checks
- Interfaced signing/unlocking of transaction inputs for easy adaptation/custimisation and extendability for any use case
- Bitcoin Transaction [Script](bscript) functionality
  - Bitcoin script engine ([interpreter](bscript/interpreter))
  - P2PKH (base58 addresses)
  - Data (OP_RETURN)
  - [BIP276](https://github.com/moneybutton/bips/blob/master/bip-0276.mediawiki)

#### Coming Soon! (18 months<sup>TM</sup>)

- Complete SigHash Flag Capability
- MultiSig functionality

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary or library deployment to Github and can be installed via: `brew install goreleaser`.

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production.
</details>

<details>
<summary><strong><code>Makefile Commands</code></strong></summary>
<br/>

View all `makefile` commands

```shell script
make help
```

List of all current commands:

```text
all                  Runs multiple commands
clean                Remove previous builds and any test cache data
clean-mods           Remove all the Go mod cache
coverage             Shows the test coverage
godocs               Sync the 