package relayer

import "github.com/cicizeo/loran/orchestrator/coingecko"

func SetPriceFeeder(pf *coingecko.PriceFeed) func(PeggyRelayer) {
	return func(s PeggyRelayer) { s.SetPriceFeeder(pf) }
}

func (s *peggyRelayer) SetPriceFeeder(pf *coingecko.PriceFeed) {
	s.priceFeeder = pf
}
