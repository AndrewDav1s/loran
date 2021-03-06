package sidechain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"

	chainclient "github.com/InjectiveLabs/sdk-go/chain/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/loran/orchestrator/ethereum/peggy"
	"github.com/InjectiveLabs/loran/orchestrator/ethereum/util"
	"github.com/InjectiveLabs/loran/orchestrator/metrics"
	"github.com/InjectiveLabs/sdk-go/chain/peggy/types"
	"github.com/InjectiveLabs/sdk-go/wrappers"
)

type PeggyBroadcastClient interface {
	FromAddress() sdk.AccAddress

	// UpdatePeggyEthAddress broadcasts a transaction updating the ETH address for the sending
	// Cosmos address. The sending Cosmos address should be a validator.
	UpdatePeggyEthAddress(
		ctx context.Context,
		ethPrivateKey *ecdsa.PrivateKey,
	) error

	// SendValsetRequest broadcasts a transaction requesting that a valset be formed for a given block
	// height.
	SendValsetRequest(
		ctx context.Context,
	) error

	// SendValsetConfirm broadcasts in a confirmation for a specific validator set for a specific block height.
	SendValsetConfirm(
		ctx context.Context,
		ethPrivateKey *ecdsa.PrivateKey,
		peggyID common.Hash,
		valset *types.Valset,
	) error

	// SendBatchConfirm broadcasts in a confirmation for a specific transaction batch set for a specific block height
	// since transaction batches also include validator sets this has all the arguments
	SendBatchConfirm(
		ctx context.Context,
		ethPrivateKey *ecdsa.PrivateKey,
		peggyID common.Hash,
		batch *types.OutgoingTxBatch,
	) error

	SendEthereumClaims(
		ctx context.Context,
		deposits []*wrappers.PeggySendToCosmosEvent,
		withdraws []*wrappers.PeggyTransactionBatchExecutedEvent,
	) error

	// SendToEth broadcasts a Tx that tokens from Cosmos to Ethereum.
	// These tokens will not be sent immediately. Instead, they will require
	// some time to be included in a batch.
	SendToEth(
		ctx context.Context,
		destination common.Address,
		amount, fee sdk.Coin,
	) error

	SendRequestBatch(
		ctx context.Context,
		denom string,
	) error
}

func NewPeggyBroadcastClient(
	queryClient types.QueryClient,
	broadcastClient chainclient.CosmosClient,
) PeggyBroadcastClient {
	return &peggyBroadcastClient{
		daemonQueryClient: queryClient,
		broadcastClient:   broadcastClient,

		svcTags: metrics.Tags{
			"svc": "peggy_broadcast",
		},
	}
}

func (s *peggyBroadcastClient) FromAddress() sdk.AccAddress {
	return s.broadcastClient.FromAddress()
}

type peggyBroadcastClient struct {
	daemonQueryClient types.QueryClient
	broadcastClient   chainclient.CosmosClient

	svcTags metrics.Tags
}

func (s *peggyBroadcastClient) UpdatePeggyEthAddress(
	ctx context.Context,
	ethPrivateKey *ecdsa.PrivateKey,
) error {
	metrics.ReportFuncCall(s.svcTags)
	doneFn := metrics.ReportFuncTiming(s.svcTags)
	defer doneFn()

	ethAddress := crypto.PubkeyToAddress(ethPrivateKey.PublicKey)
	valAddr := s.broadcastClient.FromAddress()
	// signature, err := util.NewEthereumSignature(valAddr.Bytes(), ethPrivateKey)
	signature, err := util.NewEthereumSignature(crypto.Keccak256(valAddr.Bytes()), ethPrivateKey)
	if err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.New("failed to sign validator address")
		return err
	}

	// SetEthAddress
	// This is used by the validators to set the Ethereum address that represents
	// them on the Ethereum side of the bridge. They must sign their Cosmos address
	// using the Ethereum address they have submitted. Like ValsetResponse this
	// message can in theory be submitted by anyone, but only the current validator
	// sets submissions carry any weight.
	// -------------
	msg := &types.MsgSetEthAddress{
		Address:   ethAddress.Hex(),
		Validator: valAddr.String(),
		Signature: common.Bytes2Hex(signature),
	}
	log.Infoln(s.broadcastClient)
	resp, err := s.broadcastClient.SyncBroadcastMsg(msg)
	if err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.Wrap(err, "broadcasting MsgSetEthAddress failed")
		return err
	}

	v, _ := json.Marshal(resp)
	log.Infoln("SyncBroadcastMsg resp:", string(v))
	return nil
}

func (s *peggyBroadcastClient) SendValsetRequest(
	ctx context.Context,
) error {
	metrics.ReportFuncCall(s.svcTags)
	doneFn := metrics.ReportFuncTiming(s.svcTags)
	defer doneFn()

	// ValsetRequest
	// This message starts off the validator set update process by coordinating a
	// block height around which signatures over the validators, powers, and
	// ethereum addresses will be made and submitted using a ValsetConfirm. Anyone
	// can send this message as it is not authenticated except as a valid tx. In
	// theory people could spam it and the validators will have to determine which
	// block to actually coordinate around by looking over the valset requests and
	// seeing which one some other validator has already submitted a ValsetResponse
	// for.
	// -------------
	msg := &types.MsgValsetRequest{
		Requester: s.broadcastClient.FromAddress().String(),
	}
	if err := s.broadcastClient.QueueBroadcastMsg(msg); err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.Wrap(err, "broadcasting MsgValsetRequest failed")
		return err
	}

	return nil
}

func (s *peggyBroadcastClient) SendValsetConfirm(
	ctx context.Context,
	ethPrivateKey *ecdsa.PrivateKey,
	peggyID common.Hash,
	valset *types.Valset,
) error {
	metrics.ReportFuncCall(s.svcTags)
	doneFn := metrics.ReportFuncTiming(s.svcTags)
	defer doneFn()

	ethAddress := crypto.PubkeyToAddress(ethPrivateKey.PublicKey)
	confirmHash := peggy.EncodeValsetConfirm(peggyID, valset)
	signature, err := util.NewEthereumSignature(confirmHash.Bytes(), ethPrivateKey)
	if err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.New("failed to sign validator address")
		return err
	}

	// MsgValsetConfirm
	// this is the message sent by the validators when they wish to submit their
	// signatures over the validator set at a given block height. A validator must
	// first call MsgSetEthAddress to set their Ethereum address to be used for
	// signing. Then someone (anyone) must make a ValsetRequest the request is
	// essentially a messaging mechanism to determine which block all validators
	// should submit signatures over. Finally validators sign the validator set,
	// powers, and Ethereum addresses of the entire validator set at the height of a
	// ValsetRequest and submit that signature with this message.
	//
	// If a sufficient number of validators (66% of voting power) (A) have set
	// Ethereum addresses and (B) submit ValsetConfirm messages with their
	// signatures it is then possible for anyone to view these signatures in the
	// chain store and submit them to Ethereum to update the validator set
	// -------------
	msg := &types.MsgValsetConfirm{
		Validator:  s.broadcastClient.FromAddress().String(),
		EthAddress: ethAddress.Hex(),
		Nonce:      valset.Nonce,
		Signature:  common.Bytes2Hex(signature),
	}
	if err = s.broadcastClient.QueueBroadcastMsg(msg); err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.Wrap(err, "broadcasting MsgValsetConfirm failed")
		return err
	}

	return nil
}

func (s *peggyBroadcastClient) SendBatchConfirm(
	ctx context.Context,
	ethPrivateKey *ecdsa.PrivateKey,
	peggyID common.Hash,
	batch *types.OutgoingTxBatch,
) error {
	metrics.ReportFuncCall(s.svcTags)
	doneFn := metrics.ReportFuncTiming(s.svcTags)
	defer doneFn()

	ethAddress := crypto.PubkeyToAddress(ethPrivateKey.PublicKey)
	confirmHash := peggy.EncodeTxBatchConfirm(peggyID, batch)
	signature, err := util.NewEthereumSignature(confirmHash.Bytes(), ethPrivateKey)
	if err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.New("failed to sign validator address")
		return err
	}

	// MsgConfirmBatch
	// When validators observe a MsgRequestBatch they form a batch by ordering
	// transactions currently in the txqueue in order of highest to lowest fee,
	// cutting off when the batch either reaches a hardcoded maximum size (to be
	// decided, probably around 100) or when transactions stop being profitable
	// (TODO determine this without nondeterminism) This message includes the batch
	// as well as an Ethereum signature over this batch by the validator
	// -------------
	msg := &types.MsgConfirmBatch{
		Validator:     s.broadcastClient.FromAddress().String(),
		Nonce:         batch.BatchNonce,
		Signature:     common.Bytes2Hex(signature),
		EthSigner:     ethAddress.String(),
		TokenContract: batch.TokenContract,
	}
	if err = s.broadcastClient.QueueBroadcastMsg(msg); err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.Wrap(err, "broadcasting MsgConfirmBatch failed")
		return err
	}

	return nil
}

func (s *peggyBroadcastClient) sendDepositClaims(deposit *wrappers.PeggySendToCosmosEvent) error {
	// EthereumBridgeDepositClaim
	// When more than 66% of the active validator set has
	// claimed to have seen the deposit enter the ethereum blockchain coins are
	// issued to the Cosmos address in question
	// -------------

	msg := &types.MsgDepositClaim{
		EventNonce:     deposit.EventNonce.Uint64(),
		TokenContract:  deposit.TokenContract.Hex(),
		Amount:         sdk.NewIntFromBigInt(deposit.Amount),
		EthereumSender: deposit.Sender.Hex(),
		CosmosReceiver: sdk.AccAddress(deposit.Destination[:]).String(),
		Orchestrator:   s.broadcastClient.FromAddress().String(),
		TokenName:      deposit.Name,
		TokenSymbol:    deposit.Symbol,
		TokenDecimals:  uint64(deposit.Decimals),
	}

	if err := s.broadcastClient.QueueBroadcastMsg(msg); err != nil {
		metrics.ReportFuncError(s.svcTags)
		log.WithError(err).Errorln("broadcasting MsgDepositClaim failed")
		return err
	}

	return nil
}

func (s *peggyBroadcastClient) sendWithdrawClaims(withdraw *wrappers.PeggyTransactionBatchExecutedEvent) error {
	// WithdrawClaim claims that a batch of withdrawal
	// operations on the bridge contract was executed.
	msg := &types.MsgWithdrawClaim{
		EventNonce:    withdraw.EventNonce.Uint64(),
		BatchNonce:    withdraw.BatchNonce.Uint64(),
		TokenContract: withdraw.Token.Hex(),
		Orchestrator:  s.broadcastClient.FromAddress().String(),
	}
	if err := s.broadcastClient.QueueBroadcastMsg(msg); err != nil {
		metrics.ReportFuncError(s.svcTags)
		log.WithError(err).Errorln("broadcasting MsgWithdrawClaim failed")
		return err
	}
	return nil
}

func (s *peggyBroadcastClient) SendEthereumClaims(
	ctx context.Context,
	deposits []*wrappers.PeggySendToCosmosEvent,
	withdraws []*wrappers.PeggyTransactionBatchExecutedEvent,
) error {
	metrics.ReportFuncCall(s.svcTags)
	doneFn := metrics.ReportFuncTiming(s.svcTags)
	defer doneFn()

	// Merge sort deposits, withdraws by nonce
	i, j := 0, 0
	for i < len(deposits) && j < len(withdraws) {
		if deposits[i].EventNonce.Uint64() < withdraws[j].EventNonce.Uint64() {
			if err := s.sendDepositClaims(deposits[i]); err != nil {
				metrics.ReportFuncError(s.svcTags)
				log.WithError(err).Errorln("broadcasting MsgDepositClaim failed")
				return err
			}
			i++
		} else {
			if err := s.sendWithdrawClaims(withdraws[j]); err != nil {
				metrics.ReportFuncError(s.svcTags)
				log.WithError(err).Errorln("broadcasting MsgDepositClaim failed")
				return err
			}
			j++
		}
	}

	for i < len(deposits) {
		if err := s.sendDepositClaims(deposits[i]); err != nil {
			metrics.ReportFuncError(s.svcTags)
			log.WithError(err).Errorln("broadcasting MsgDepositClaim failed")
			return err
		}
		i++
	}

	for j < len(withdraws) {
		if err := s.sendWithdrawClaims(withdraws[j]); err != nil {
			metrics.ReportFuncError(s.svcTags)
			log.WithError(err).Errorln("broadcasting MsgDepositClaim failed")
			return err
		}
		j++
	}
	return nil
}

func (s *peggyBroadcastClient) SendToEth(
	ctx context.Context,
	destination common.Address,
	amount, fee sdk.Coin,
) error {
	metrics.ReportFuncCall(s.svcTags)
	doneFn := metrics.ReportFuncTiming(s.svcTags)
	defer doneFn()

	// MsgSendToEth
	// This is the message that a user calls when they want to bridge an asset
	// it will later be removed when it is included in a batch and successfully
	// submitted tokens are removed from the users balance immediately
	// -------------
	// AMOUNT:
	// the coin to send across the bridge, note the restriction that this is a
	// single coin not a set of coins that is normal in other Cosmos messages
	// FEE:
	// the fee paid for the bridge, distinct from the fee paid to the chain to
	// actually send this message in the first place. So a successful send has
	// two layers of fees for the user
	msg := &types.MsgSendToEth{
		Sender:    s.broadcastClient.FromAddress().String(),
		EthDest:   destination.Hex(),
		Amount:    amount,
		BridgeFee: fee, // TODO: use exactly that fee for transaction
	}
	if err := s.broadcastClient.QueueBroadcastMsg(msg); err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.Wrap(err, "broadcasting MsgSendToEth failed")
		return err
	}

	return nil
}

func (s *peggyBroadcastClient) SendRequestBatch(
	ctx context.Context,
	denom string,
) error {
	metrics.ReportFuncCall(s.svcTags)
	doneFn := metrics.ReportFuncTiming(s.svcTags)
	defer doneFn()

	// MsgRequestBatch
	// this is a message anyone can send that requests a batch of transactions to
	// send across the bridge be created for whatever block height this message is
	// included in. This acts as a coordination point, the handler for this message
	// looks at the AddToOutgoingPool tx's in the store and generates a batch, also
	// available in the store tied to this message. The validators then grab this
	// batch, sign it, submit the signatures with a MsgConfirmBatch before a relayer
	// can finally submit the batch
	// -------------
	msg := &types.MsgRequestBatch{
		Denom:     denom,
		Requester: s.broadcastClient.FromAddress().String(),
	}
	if err := s.broadcastClient.QueueBroadcastMsg(msg); err != nil {
		metrics.ReportFuncError(s.svcTags)
		err = errors.Wrap(err, "broadcasting MsgRequestBatch failed")
		return err
	}

	return nil
}
