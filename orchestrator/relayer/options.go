package relayer

import "github.com/cicizeo/loran/orchestrator/coingecko"

func SetPriceFeeder(pf *coingecko.PriceFeed) func(GravityRelayer) {
	return func(s GravityRelayer) { s.SetPriceFeeder(pf) }
}

func (s *gravityRelayer) SetPriceFeeder(pf *coingecko.PriceFeed) {
	s.priceFeeder = pf
}
