package orchestrator

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/pkg/errors"

	wrappers "github.com/cicizeo/loran/solidity/wrappers/Peggy.sol"
	"github.com/cicizeo/hilo/x/peggy/types"
)

// CheckForEvents checks for events such as a deposit to the Peggy Ethereum contract or a validator set update
// or a transaction batch update. It then responds to these events by performing actions on the Cosmos chain if required
func (p *peggyOrchestrator) CheckForEvents(
	ctx context.Context,
	startingBlock uint64,
	ethBlockConfirmationDelay uint64,
) (currentBlock uint64, err error) {

	latestHeader, err := p.ethProvider.HeaderByNumber(ctx, nil)
	if err != nil {
		err = errors.Wrap(err, "failed to get latest header")
		return 0, err
	}

	// add delay to ensure minimum confirmations are received and block is finalized
	currentBlock = latestHeader.Number.Uint64() - ethBlockConfirmationDelay

	if currentBlock < startingBlock {
		return currentBlock, nil
	}

	if (currentBlock - startingBlock) > p.ethBlocksPerLoop {
		currentBlock = startingBlock + p.ethBlocksPerLoop
	}

	peggyFilterer, err := wrappers.NewPeggyFilterer(p.peggyContract.Address(), p.ethProvider)
	if err != nil {
		err = errors.Wrap(err, "failed to init Peggy events filterer")
		return 0, err
	}

	var erc20DeployedEvents []*wrappers.PeggyERC20DeployedEvent
	{
		iter, err := peggyFilterer.FilterERC20DeployedEvent(&bind.FilterOpts{
			Start: startingBlock,
			End:   &currentBlock,
		}, nil)
		if err != nil {
			p.logger.Err(err).
				Uint64("start", startingBlock).
				Uint64("end", currentBlock).
				Msg("failed to scan past ERC20Deployed events from Ethereum")

			if !isUnknownBlockErr(err) {
				err = errors.Wrap(err, "failed to scan past ERC20Deployed events from Ethereum")
				return 0, err
			} else if iter == nil {
				return 0, errors.New("no iterator returned")
			}
		}

		for iter.Next() {
			erc20DeployedEvents = append(erc20DeployedEvents, iter.Event)
		}

		iter.Close()
	}

	p.logger.Debug().
		Uint64("start", startingBlock).
		Uint64("end", currentBlock).
		Int("num_events", len(erc20DeployedEvents)).
		Msg("scanned ERC20Deployed events from Ethereum")

	var sendToCosmosEvents []*wrappers.PeggySendToCosmosEvent
	{

		iter, err := peggyFilterer.FilterSendToCosmosEvent(&bind.FilterOpts{
			Start: startingBlock,
			End:   &currentBlock,
		}, nil, nil, nil)
		if err != nil {
			p.logger.Err(err).
				Uint64("start", startingBlock).
				Uint64("end", currentBlock).
				Msg("failed to scan past SendToCosmos events from Ethereum")

			if !isUnknownBlockErr(err) {
				err = errors.Wrap(err, "failed to scan past SendToCosmos events from Ethereum")
				return 0, err
			} else if iter == nil {
				return 0, errors.New("no iterator returned")
			}
		}

		for iter.Next() {
			sendToCosmosEvents = append(sendToCosmosEvents, iter.Event)
		}

		iter.Close()
	}

	p.logger.Debug().
		Uint64("start", startingBlock).
		Uint64("end", currentBlock).
		Int("num_events", len(sendToCosmosEvents)).
		Msg("scanned SendToCosmos events from Ethereum")

	var transactionBatchExecutedEvents []*wrappers.PeggyTransactionBatchExecutedEvent
	{
		iter, err := peggyFilterer.FilterTransactionBatchExecutedEvent(&bind.FilterOpts{
			Start: startingBlock,
			End:   &currentBlock,
		}, nil, nil)
		if err != nil {
			p.logger.Err(err).
				Uint64("start", startingBlock).
				Uint64("end", currentBlock).
				Msg("failed to scan past TransactionBatchExecuted events from Ethereum")

			if !isUnknownBlockErr(err) {
				err = errors.Wrap(err, "failed to scan past TransactionBatchExecuted events from Ethereum")
				return 0, err
			} else if iter == nil {
				return 0, errors.New("no iterator returned")
			}
		}

		for iter.Next() {
			transactionBatchExecutedEvents = append(transactionBatchExecutedEvents, iter.Event)
		}

		iter.Close()
	}

	p.logger.Debug().
		Uint64("start", startingBlock).
		Uint64("end", currentBlock).
		Int("num_events", len(transactionBatchExecutedEvents)).
		Msg("scanned TransactionBatchExecuted events from Ethereum")

	var valsetUpdatedEvents []*wrappers.PeggyValsetUpdatedEvent
	{
		iter, err := peggyFilterer.FilterValsetUpdatedEvent(&bind.FilterOpts{
			Start: startingBlock,
			End:   &currentBlock,
		}, nil)
		if err != nil {
			p.logger.Err(err).
				Uint64("start", startingBlock).
				Uint64("end", currentBlock).
				Msg("failed to scan past ValsetUpdatedEvent events from Ethereum")

			if !isUnknownBlockErr(err) {
				err = errors.Wrap(err, "failed to scan past ValsetUpdatedEvent events from Ethereum")
				return 0, err
			} else if iter == nil {
				return 0, errors.New("no iterator returned")
			}
		}

		for iter.Next() {
			valsetUpdatedEvents = append(valsetUpdatedEvents, iter.Event)
		}

		iter.Close()
	}

	p.logger.Debug().
		Uint64("start", startingBlock).
		Uint64("end", currentBlock).
		Int("num_events", len(valsetUpdatedEvents)).
		Msg("scanned ValsetUpdatedEvents events from Ethereum")

	// note that starting block overlaps with our last checked block, because we have to deal with
	// the possibility that the relayer was killed after relaying only one of multiple events in a single
	// block, so we also need this routine so make sure we don't send in the first event in this hypothetical
	// multi event block again. In theory we only send all events for every block and that will pass of fail
	// atomically but lets not take that risk.
	lastEventResp, err := p.cosmosQueryClient.LastEventByAddr(ctx, &types.QueryLastEventByAddrRequest{
		Address: p.peggyBroadcastClient.AccFromAddress().String(),
	})

	if err != nil {
		err = errors.New("failed to query last claim event from backend")
		return 0, err
	}

	if lastEventResp == nil {
		return 0, errors.New("no last event response returned")
	}

	lastClaimEvent := lastEventResp.LastClaimEvent

	deposits := filterSendToCosmosEventsByNonce(sendToCosmosEvents, lastClaimEvent.EthereumEventNonce)
	withdraws := filterTransactionBatchExecutedEventsByNonce(
		transactionBatchExecutedEvents,
		lastClaimEvent.EthereumEventNonce,
	)
	valsetUpdates := filterValsetUpdateEventsByNonce(valsetUpdatedEvents, lastClaimEvent.EthereumEventNonce)
	deployedERC20Updates := filterERC20DeployedEventsByNonce(erc20DeployedEvents, lastClaimEvent.EthereumEventNonce)

	if len(deposits) > 0 || len(withdraws) > 0 || len(valsetUpdates) > 0 || len(deployedERC20Updates) > 0 {

		if err := p.peggyBroadcastClient.SendEthereumClaims(
			ctx,
			lastClaimEvent.EthereumEventNonce,
			deposits,
			withdraws,
			valsetUpdates,
			deployedERC20Updates,
			p.cosmosBlockTime,
		); err != nil {
			err = errors.Wrap(err, "failed to send ethereum claims to Cosmos chain")
			return 0, err
		}
	}

	return currentBlock, nil
}

func filterSendToCosmosEventsByNonce(
	events []*wrappers.PeggySendToCosmosEvent,
	nonce uint64,
) []*wrappers.PeggySendToCosmosEvent {
	res := make([]*wrappers.PeggySendToCosmosEvent, 0, len(events))

	for _, ev := range events {
		if ev.EventNonce.Uint64() > nonce {
			res = append(res, ev)
		}
	}

	return res
}

func filterTransactionBatchExecutedEventsByNonce(
	events []*wrappers.PeggyTransactionBatchExecutedEvent,
	nonce uint64,
) []*wrappers.PeggyTransactionBatchExecutedEvent {
	res := make([]*wrappers.PeggyTransactionBatchExecutedEvent, 0, len(events))

	for _, ev := range events {
		if ev.EventNonce.Uint64() > nonce {
			res = append(res, ev)
		}
	}

	return res
}

func filterValsetUpdateEventsByNonce(
	events []*wrappers.PeggyValsetUpdatedEvent,
	nonce uint64,
) []*wrappers.PeggyValsetUpdatedEvent {
	res := make([]*wrappers.PeggyValsetUpdatedEvent, 0, len(events))

	for _, ev := range events {
		if ev.EventNonce.Uint64() > nonce {
			res = append(res, ev)
		}
	}
	return res
}

func filterERC20DeployedEventsByNonce(
	events []*wrappers.PeggyERC20DeployedEvent,
	nonce uint64,
) []*wrappers.PeggyERC20DeployedEvent {
	res := make([]*wrappers.PeggyERC20DeployedEvent, 0, len(events))

	for _, ev := range events {
		if ev.EventNonce.Uint64() > nonce {
			res = append(res, ev)
		}
	}
	return res
}

func isUnknownBlockErr(err error) bool {
	// Geth error
	if strings.Contains(err.Error(), "unknown block") {
		return true
	}

	// Parity error
	if strings.Contains(err.Error(), "One of the blocks specified in filter") {
		return true
	}

	return false
}
