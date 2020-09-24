# Loran

<!-- markdownlint-disable MD041 -->

[![Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://img.shields.io/badge/repo%20status-WIP-yellow.svg?style=flat-square)](https://www.repostatus.org/#wip)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/cicizeo/loran)
[![Go Report Card](https://goreportcard.com/badge/github.com/cicizeo/loran?style=flat-square)](https://goreportcard.com/report/github.com/cicizeo/loran)
[![Version](https://img.shields.io/github/tag/cicizeo/loran.svg?style=flat-square)](https://github.com/cicizeo/loran/releases/latest)
[![License: Apache-2.0](https://img.shields.io/github/license/cicizeo/loran.svg?style=flat-square)](https://github.com/cicizeo/loran/blob/main/LICENSE)
[![Lines Of Code](https://img.shields.io/tokei/lines/github/cicizeo/loran?style=flat-square)](https://github.com/cicizeo/loran)
[![GitHub Super-Linter](https://img.shields.io/github/workflow/status/cicizeo/loran/Lint?style=flat-square&label=Lint)](https://github.com/marketplace/actions/super-linter)

Loran is a Go implementation of the Peggy (Gravity Bridge) Orchestrator originally
implemented by [Injective Labs](https://github.com/InjectiveLabs/). Loran itself
is a fork of the original Gravity Bridge Orchestrator implemented by [Althea](https://github.com/althea-net).


## Table of Contents

- [Dependencies](#dependencies)
- [Installation](#installation)
- [How to run](#how-to-run)
- [How it works](#how-it-works)

## Dependencies

- [Go 1.17+](https://golang.org/dl/)

## Installation

To install the `loran` binary:

```shell
$ make install
```

## How to run

### Setup

First we must register the validator's Ethereum key. This key will be used to
sign claims going from Ethereum to Hilo and to sign any transactions sent to
Ethereum (batches or validator set updates).

```shell
$ loran tx register-eth-key \
  --cosmos-chain-id="..." \
  --cosmos-grpc="tcp://..." \
  --tendermint-rpc="http://..." \
  --cosmos-keyring=... \
  --cosmos-keyring-dir=... \
  --cosmos-from=... \
  --eth-pk=$ETH_PK
```

### Run the orchestrator

```shell
$ loran orchestrator \
  --eth-pk=$ETH_PK \
  --eth-rpc=$ETH_RPC \
  --relay-batches=true \
  --relay-valsets=true \
  --eth-chain-id=... \
  --cosmos-chain-id=... \
  --cosmos-grpc="tcp://..." \
  --tendermint-rpc="http://..." \
  --cosmos-keyring=... \
  --cosmos-keyring-dir=... \
  --cosmos-from=...
```

### Send a transfer from Hilo to Ethereum

This is done using the command `hilod tx peggy send-to-eth`, use the `--help`
flag for more information.

If the coin doesn't have a corresponding ERC20 equivalent on the Ethereum
network, the transaction will fail. This is only required for Cosmos originated
coins and anyone can call the `deployERC20` function on the Peggy contract to
fix this (Loran has a helper command for this, see
`loran bridge deploy-erc20 --help` for more details).

This process takes longer than transfers the other way around because they get
relayed in batches rather than individually. It primarily depends on the amount
of transfers of the same token and the fees the senders are paying.

Important notice: if an "unlisted" (with no monetary value) ERC20 token gets
sent into Hilo it won't be possible to transfer it back to Ethereum, unless a
validator is configured to batch and relay transactions of this token.

### Send a transfer from Ethereum to Hilo

Any ERC20 token can be sent to Hilo and it's done using the command
`loran bridge send-to-cosmos`, use the `--help` flag for more information. It
can also be done by calling the `sendToCosmos` method on the Peggy contract.

The ERC20 tokens will be locked in the Peggy contract and new coins will be
minted on Hilo with the denomination `peggy{token_address}`. This process takes
around 3 minutes or 12 Ethereum blocks.

## How it works

Loran allows transfers of assets back and forth between Ethereum and Hilo.
It supports both assets originating on Hilo and assets originating on Ethereum
(any ERC20 token).

It works by scanning the events of the contract deployed on Ethereum (Peggy) and
relaying them as messages to the Hilo chain; and relaying transaction batches and
validator sets from Hilo to Ethereum.

### Events and messages observed/relayed

#### Ethereum

**Deposits** (`SendToCosmosEvent`): emitted when sending tokens from Ethereum to
Hilo using the `sendToCosmos` function on Peggy.

**Withdraw** (`TransactionBatchExecutedEvent`): emitted when a batch of
transactions is sent from Hilo to Ethereum using the `submitBatch` function on
the Peggy contract by a validator. This serves as a confirmation to Hilo that
the batch was sent successfully.

**Valset update** (`ValsetUpdatedEvent`): emitted on init of the Peggy contract
and on every execution of the `updateValset` function.

**Deployed ERC 20** (`ERC20DeployedEvent`): emitted when executing the function
`deployERC20`. This event signals Hilo that there's a new ERC20 deployed from
Peggy, so Hilo can map the token contract address to the corresponding native
coin. This enables transfers from Hilo to Ethereum.

#### Hilo

 **Validator sets**: Hilo informs the Peggy contract who are the current
 validators and their power. This results in an execution of the `updateValset`
 function.

 **Request batch**: Loran will check for new transactions in the Outgoing TX Pool
 and if the transactions' fees are greater than the set minimum batch fee, it
 will send a message to Hilo requesting a new batch.

 **Batches**: Loran queries Hilo for any batches ready to be relayed and relays
 them over to Ethereum using the `submitBatch` function on the Peggy contract.
