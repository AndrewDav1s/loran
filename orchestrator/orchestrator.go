package orchestrator

import (
	"context"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	"github.com/cicizeo/loran/orchestrator/coingecko"
	"github.com/cicizeo/loran/orchestrator/cosmos/tmclient"

	sidechain "github.com/cicizeo/loran/orchestrator/cosmos"
	"github.com/cicizeo/loran/orchestrator/ethereum/keystore"
	"github.com/cicizeo/loran/orchestrator/ethereum/peggy"
	"github.com/cicizeo/loran/orchestrator/ethereum/provider"
	"github.com/cicizeo/loran/orchestrator/relayer"
)

type PeggyOrchestrator interface {
	Start(ctx context.Context) error
	CheckForEvents(ctx context.Context, startingBlock uint64) (currentBlock uint64, err error)
	GetLastCheckedBlock(ctx context.Context) (uint64, error)
	EthOracleMainLoop(ctx context.Context) error
	EthSignerMainLoop(ctx context.Context) error
	BatchRequesterLoop(ctx context.Context) error
	RelayerMainLoop(ctx context.Context) error

	SetMinBatchFee(float64)
	SetPriceFeeder(*coingecko.CoingeckoPriceFeed)
}

type peggyOrchestrator struct {
	logger               zerolog.Logger
	tmClient             tmclient.TendermintClient
	cosmosQueryClient    sidechain.PeggyQueryClient
	peggyBroadcastClient sidechain.PeggyBroadcastClient
	peggyContract        peggy.PeggyContract
	ethProvider          provider.EVMProvider
	ethFrom              ethcmn.Address
	ethSignerFn          keystore.SignerFn
	ethPersonalSignFn    keystore.PersonalSignFn
	relayer              relayer.PeggyRelayer

	// optional inputs with defaults
	minBatchFeeUSD float64
	priceFeeder    *coingecko.CoingeckoPriceFeed
}

func NewPeggyOrchestrator(
	logger zerolog.Logger,
	cosmosQueryClient sidechain.PeggyQueryClient,
	peggyBroadcastClient sidechain.PeggyBroadcastClient,
	tmClient tmclient.TendermintClient,
	peggyContract peggy.PeggyContract,
	ethFrom ethcmn.Address,
	ethSignerFn keystore.SignerFn,
	ethPersonalSignFn keystore.PersonalSignFn,
	relayer relayer.PeggyRelayer,
	options ...func(PeggyOrchestrator),
) PeggyOrchestrator {

	var orch PeggyOrchestrator
	orch = &peggyOrchestrator{
		logger:               logger.With().Str("module", "orchestrator").Logger(),
		tmClient:             tmClient,
		cosmosQueryClient:    cosmosQueryClient,
		peggyBroadcastClient: peggyBroadcastClient,
		peggyContract:        peggyContract,
		ethProvider:          peggyContract.Provider(),
		ethFrom:              ethFrom,
		ethSignerFn:          ethSignerFn,
		ethPersonalSignFn:    ethPersonalSignFn,
		relayer:              relayer,
	}

	for _, option := range options {
		option(orch)
	}

	return orch
}
