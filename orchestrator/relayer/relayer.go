package relayer

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/cicizeo/loran/orchestrator/coingecko"
	"github.com/cicizeo/loran/orchestrator/cosmos"
	"github.com/cicizeo/loran/orchestrator/cosmos/tmclient"
	"github.com/cicizeo/loran/orchestrator/ethereum/peggy"
	"github.com/cicizeo/loran/orchestrator/ethereum/provider"
	"github.com/cicizeo/hilo/x/peggy/types"
)

type PeggyRelayer interface {
	Start(ctx context.Context) error

	FindLatestValset(ctx context.Context) (*types.Valset, error)
	RelayBatches(
		ctx context.Context,
		currentValset *types.Valset,
		possibleBatches map[common.Address][]SubmittableBatch,
	) error
	RelayValsets(ctx context.Context, currentValset *types.Valset) error

	SetPriceFeeder(*coingecko.PriceFeed)
}

type peggyRelayer struct {
	logger             zerolog.Logger
	cosmosQueryClient  cosmos.PeggyQueryClient
	peggyContract      peggy.Contract
	ethProvider        provider.EVMProvider
	tmClient           tmclient.TendermintClient
	valsetRelayEnabled bool
	batchRelayEnabled  bool
	ethereumBlockTime  time.Duration
	priceFeeder        *coingecko.PriceFeed

	// store locally the last tx this validator made to avoid sending duplicates
	// or invalid txs
	lastSentBatchNonce uint64
}

func NewPeggyRelayer(
	logger zerolog.Logger,
	cosmosQueryClient cosmos.PeggyQueryClient,
	peggyContract peggy.Contract,
	tmClient tmclient.TendermintClient,
	valsetRelayEnabled bool,
	batchRelayEnabled bool,
	ethereumBlockTime time.Duration,
	options ...func(PeggyRelayer),
) PeggyRelayer {
	relayer := &peggyRelayer{
		logger:             logger.With().Str("module", "peggy_relayer").Logger(),
		cosmosQueryClient:  cosmosQueryClient,
		peggyContract:      peggyContract,
		tmClient:           tmClient,
		ethProvider:        peggyContract.Provider(),
		valsetRelayEnabled: valsetRelayEnabled,
		batchRelayEnabled:  batchRelayEnabled,
		ethereumBlockTime:  ethereumBlockTime,
	}

	for _, option := range options {
		option(relayer)
	}

	return relayer
}
