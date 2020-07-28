package relayer

import (
	"context"

	"github.com/InjectiveLabs/sdk-go/chain/peggy/types"
	"github.com/cicizeo/loran/orchestrator/cosmos"
	"github.com/cicizeo/loran/orchestrator/ethereum/peggy"
	"github.com/cicizeo/loran/orchestrator/ethereum/provider"
	"github.com/cicizeo/loran/orchestrator/metrics"
)

type PeggyRelayer interface {
	Start(ctx context.Context) error

	FindLatestValset(ctx context.Context) (*types.Valset, error)
	RelayBatches(ctx context.Context) error
	RelayValsets(ctx context.Context) error
}

type peggyRelayer struct {
	svcTags metrics.Tags

	cosmosQueryClient  cosmos.PeggyQueryClient
	peggyContract      peggy.PeggyContract
	ethProvider        provider.EVMProvider
	valsetRelayEnabled bool
	batchRelayEnabled  bool
}

func NewPeggyRelayer(
	cosmosQueryClient cosmos.PeggyQueryClient,
	peggyContract peggy.PeggyContract,
	valsetRelayEnabled bool,
	batchRelayEnabled bool,
) PeggyRelayer {
	return &peggyRelayer{
		cosmosQueryClient:  cosmosQueryClient,
		peggyContract:      peggyContract,
		ethProvider:        peggyContract.Provider(),
		valsetRelayEnabled: valsetRelayEnabled,
		batchRelayEnabled:  batchRelayEnabled,
		svcTags: metrics.Tags{
			"svc": "peggy_relayer",
		},
	}
}
