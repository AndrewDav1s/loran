package orchestrator

import (
	"context"
	"crypto/ecdsa"

	"github.com/InjectiveLabs/loran/orchestrator/ethereum/peggy"
	"github.com/InjectiveLabs/loran/orchestrator/ethereum/provider"
	"github.com/InjectiveLabs/loran/orchestrator/metrics"
	"github.com/InjectiveLabs/loran/orchestrator/sidechain"
	"github.com/InjectiveLabs/loran/orchestrator/sidechain/tmclient"
)

type PeggyOrchestrator interface {
	RunLoop(ctx context.Context)
}

type peggyOrchestrator struct {
	svcTags metrics.Tags

	tmClient             tmclient.TendermintClient
	cosmosQueryClient    sidechain.PeggyQueryClient
	peggyBroadcastClient sidechain.PeggyBroadcastClient
	peggyContract        peggy.PeggyContract
	ethProvider          provider.EVMProvider
	ethPrivateKey        *ecdsa.PrivateKey
	injContractAddress   string
}

func NewPeggyOrchestrator(
	cosmosQueryClient sidechain.PeggyQueryClient,
	peggyBroadcastClient sidechain.PeggyBroadcastClient,
	tmClient tmclient.TendermintClient,
	peggyContract peggy.PeggyContract,
	ethPrivateKey *ecdsa.PrivateKey,
	injContractAddress string,
) PeggyOrchestrator {
	return &peggyOrchestrator{
		tmClient:             tmClient,
		cosmosQueryClient:    cosmosQueryClient,
		peggyBroadcastClient: peggyBroadcastClient,
		peggyContract:        peggyContract,
		ethProvider:          peggyContract.Provider(),
		ethPrivateKey:        ethPrivateKey,
		injContractAddress:   injContractAddress,

		svcTags: metrics.Tags{
			"svc": "peggy_orchestrator",
		},
	}
}
