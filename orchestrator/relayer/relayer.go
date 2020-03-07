package relayer

import (
	"context"

	"github.com/InjectiveLabs/loran/modules/peggy/types"
	"github.com/InjectiveLabs/loran/orchestrator/cosmos"
	"github.com/InjectiveLabs/loran/orchestrator/ethereum/peggy"
	"github.com/InjectiveLabs/loran/orchestrator/ethereum/provider"
	"github.com/InjectiveLabs/loran/orchestrator/metrics"
)

type PeggyRelayer interface {
	Start(ctx context.Context) error

	FindLatestValset(ctx context.Context) (*types.Valset, error)
	RelayBatches(ctx context.Context) error
	RelayValsets(ctx context.Context) error
}

type peggyRelayer struct {
	svcTags metrics.Tags

	cosmosQueryClient cosmos.PeggyQueryClient
	peggyContract     peggy.PeggyContract
	ethProvider       provider.EVMProvider
}

func NewPeggyRelayer(
	cosmosQueryClient cosmos.PeggyQueryClient,
	peggyContract peggy.PeggyContract,
) PeggyRelayer {
	return &peggyRelayer{
		cosmosQueryClient: cosmosQueryClient,
		peggyContract:     peggyContract,
		ethProvider:       peggyContract.Provider(),

		svcTags: metrics.Tags{
			"svc": "peggy_relayer",
		},
	}
}
