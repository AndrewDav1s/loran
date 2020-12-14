package relayer

import (
	"context"

	retry "github.com/avast/retry-go"
	"github.com/pkg/errors"

	"github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	"github.com/cicizeo/loran/orchestrator/loops"
)

func (s *gravityRelayer) Start(ctx context.Context) error {
	logger := s.logger.With().Str("loop", "RelayerMainLoop").Logger()

	if s.valsetRelayEnabled {
		logger.Info().Msg("valset relay enabled; starting to relay valsets to Ethereum")
	}

	if s.batchRelayEnabled {
		logger.Info().Msg("batch relay enabled; starting to relay batches to Ethereum")
	}

	return loops.RunLoop(ctx, s.logger, s.loopDuration, func() error {
		var (
			currentValset *types.Valset
			err           error
		)

		err = retry.Do(func() error {
			currentValset, err = s.FindLatestValset(ctx)
			if err != nil {
				return errors.New("failed to find latest valset")
			}

			return nil
		}, retry.Context(ctx), retry.OnRetry(func(n uint, err error) {
			logger.Err(err).Uint("retry", n).Msg("failed to find latest valset; retrying...")
		}))

		if err != nil {
			s.logger.Panic().Err(err).Msg("exhausted retries to get latest valset")
		}

		var pg loops.ParanoidGroup
		if s.valsetRelayEnabled {
			pg.Go(func() error {
				return retry.Do(func() error {
					return s.RelayValsets(ctx, *currentValset)
				}, retry.Context(ctx), retry.OnRetry(func(n uint, err error) {
					logger.Err(err).Uint("retry", n).Msg("failed to relay valsets; retrying...")
				}))
			})
		}

		if s.batchRelayEnabled {
			pg.Go(func() error {
				return retry.Do(func() error {

					possibleBatches, err := s.getBatchesAndSignatures(ctx, *currentValset)
					if err != nil {
						return err
					}

					return s.RelayBatches(ctx, *currentValset, possibleBatches)
				}, retry.Context(ctx), retry.OnRetry(func(n uint, err error) {
					logger.Err(err).Uint("retry", n).Msg("failed to relay tx batches; retrying...")
				}))
			})
		}

		if pg.Initialized() {
			if err := pg.Wait(); err != nil {
				logger.Err(err).Msg("main relay loop failed; exiting...")
				return err
			}
		}
		return nil
	})
}
