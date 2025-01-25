package strategy

import (
	"github.com/sachin-404/gobalancer/pkg/domain"
	log "github.com/sirupsen/logrus"
	"sync"
)

const (
	RoundRobinStrategy         = "RoundRobin"
	WeightedRoundRobinStrategy = "WeightedRoundRobin"
	UnknownStrategy            = "Unknown"
)

// BalancingStrategy is the load balancing strategy that every algorithm should implement
type BalancingStrategy interface {
	Next([]*domain.Server) (*domain.Server, error)
}

var strategies = map[string]func() BalancingStrategy{}

func init() {
	strategies = make(map[string]func() BalancingStrategy)

	strategies[RoundRobinStrategy] = func() BalancingStrategy {
		return &RoundRobin{
			mu:      sync.Mutex{},
			Current: 0,
		}
	}

	strategies[WeightedRoundRobinStrategy] = func() BalancingStrategy {
		return &WeightedRoundRobin{mu: sync.Mutex{}}
	}
}

// LoadStrategy will try to resolve the strategy name
// default to RoundRobinStrategy if no strategy matched
func LoadStrategy(strategyName string) BalancingStrategy {
	strategy, ok := strategies[strategyName]
	if !ok {
		// Default if provided strategy does not match
		log.Warnf("Strategy '%s' not found, falling back to RoundRobin strategy", strategyName)
		return strategies[RoundRobinStrategy]()
	}
	return strategy()
}
