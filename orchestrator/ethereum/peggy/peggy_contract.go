package peggy

import (
	"context"
	"math"
	"math/big"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/cicizeo/loran/orchestrator/ethereum/committer"
	"github.com/cicizeo/loran/orchestrator/ethereum/provider"
	wrappers "github.com/cicizeo/loran/solidity/wrappers/Peggy.sol"
	"github.com/cicizeo/hilo/x/peggy/types"
)

type Contract interface {
	committer.EVMCommitter

	// Address returns the Peggy contract address
	Address() common.Address

	// SendToCosmos executes the function of the same name on the Peggy contract to send ERC20 tokens to Cosmos.
	SendToCosmos(
		ctx context.Context,
		erc20 common.Address,
		amount *big.Int,
		cosmosAccAddress sdk.AccAddress,
		senderAddress common.Address,
	) (*common.Hash, error)

	// EncodeTransactionBatch encodes a batch into a tx byte data. This is specially helpful for estimating gas and
	// detecting identical transactions in the mempool.
	EncodeTransactionBatch(
		ctx context.Context,
		currentValset *types.Valset,
		batch *types.OutgoingTxBatch,
		confirms []*types.MsgConfirmBatch,
	) ([]byte, error)

	// EncodeValsetUpdate encodes a valset update into a tx byte data. This is specially helpful for estimating gas and
	// detecting identical transactions in the mempool.
	EncodeValsetUpdate(
		ctx context.Context,
		oldValset *types.Valset,
		newValset *types.Valset,
		confirms []*types.MsgValsetConfirm,
	) ([]byte, error)

	GetTxBatchNonce(
		ctx context.Context,
		erc20ContractAddress common.Address,
		callerAddress common.Address,
	) (*big.Int, error)

	GetValsetNonce(
		ctx context.Context,
		callerAddress common.Address,
	) (*big.Int, error)

	GetPeggyID(
		ctx context.Context,
		callerAddress common.Address,
	) (common.Hash, error)

	GetERC20Symbol(
		ctx context.Context,
		erc20ContractAddress common.Address,
		callerAddress common.Address,
	) (symbol string, err error)

	GetERC20Decimals(
		ctx context.Context,
		erc20ContractAddress common.Address,
		callerAddress common.Address,
	) (decimals uint8, err error)

	// SubscribeToPendingTxs starts a websocket connection to Alchemy's service that listens for new pending txs made
	// to the Peggy contract.
	SubscribeToPendingTxs(ctx context.Context, alchemyWebsocketURL string) error

	// IsPendingTxInput returns true if the input data is found in the pending tx list. If the tx is found but the tx is
	// older than pendingTxWaitDuration, we consider it stale and return false, so the validator re-sends it.
	IsPendingTxInput(txData []byte, pendingTxWaitDuration time.Duration) bool
}

func NewPeggyContract(
	logger zerolog.Logger,
	ethCommitter committer.EVMCommitter,
	peggyAddress common.Address,
) (Contract, error) {
	ethPeggy, err := wrappers.NewPeggy(peggyAddress, ethCommitter.Provider())
	if err != nil {
		return nil, err
	}

	svc := &peggyContract{
		logger:       logger.With().Str("module", "peggy_contract").Logger(),
		EVMCommitter: ethCommitter,
		peggyAddress: peggyAddress,
		ethPeggy:     ethPeggy,
	}

	return svc, nil
}

type peggyContract struct {
	logger zerolog.Logger
	committer.EVMCommitter

	ethProvider  provider.EVMProvider
	peggyAddress common.Address
	ethPeggy     *wrappers.Peggy

	pendingTxInputList PendingTxInputList
}

func (s *peggyContract) Address() common.Address {
	return s.peggyAddress
}

// maxUintAllowance is uint constant MAX_UINT = 2**256 - 1
var maxUintAllowance = big.NewInt(0).Sub(big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))

var (
	peggyABI, _ = abi.JSON(strings.NewReader(wrappers.PeggyABI))
	erc20ABI, _ = abi.JSON(strings.NewReader(wrappers.ERC20ABI))
)

func sigToVRS(sigHex string) (v uint8, r, s common.Hash) {
	signatureBytes := common.FromHex(sigHex)
	vParam := signatureBytes[64]
	if vParam == byte(0) {
		vParam = byte(27)
	} else if vParam == byte(1) {
		vParam = byte(28)
	}

	v = vParam
	r = common.BytesToHash(signatureBytes[0:32])
	s = common.BytesToHash(signatureBytes[32:64])

	return
}

// The total power in the Peggy bridge is normalized to u32 max every
// time a validator set is created. This value of up to u32 max is then
// stored in a i64 to prevent overflow during computation.
const totalPeggyPower int64 = math.MaxUint32

// peggyPowerToPercent takes in an amount of power in the peggy bridge, returns a percentage of total
func peggyPowerToPercent(total *big.Int) float32 {
	d := decimal.NewFromBigInt(total, 0)
	f, _ := d.Div(decimal.NewFromInt(totalPeggyPower)).Shift(2).Float64()
	return float32(f)
}

var ErrInsufficientVotingPowerToPass = errors.New("insufficient voting power")
