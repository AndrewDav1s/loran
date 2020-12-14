package relayer

import (
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/cicizeo/loran/mocks"
	gravityMocks "github.com/cicizeo/loran/mocks/gravity"
	"github.com/cicizeo/loran/orchestrator/ethereum/provider"
)

func TestNewGravityRelayer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	mockQClient := mocks.NewMockQueryClient(mockCtrl)
	mockGravityContract := gravityMocks.NewMockContract(mockCtrl)
	mockGravityContract.EXPECT().Provider().Return(provider.NewEVMProvider(nil))

	relayer := NewGravityRelayer(logger,
		mockQClient,
		mockGravityContract,
		true,
		true,
		time.Minute,
		time.Minute,
		1.0,
		SetPriceFeeder(nil),
	)

	assert.NotNil(t, relayer)
}
