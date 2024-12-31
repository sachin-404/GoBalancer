package strategy

import (
	"github.com/sachin-404/gobalancer/pkg/domain"
	"sync/atomic"
)

const (
	RoundRobinStrategy         = "RoundRobin"
	WeightedRoundRobinStrategy = "WeightedRoundRobin"
	UnknownStrategy            = "Unknown"
)

// BalancingStrategy is the loadbalancing strategy that every algorithm should implement
type BalancingStrategy interface {
	Next([]*domain.Server) (*domain.Server, error)
}

var strategies = map[string]func() BalancingStrategy{}

func init() {
	strategies = make(map[string]func() BalancingStrategy)
	strategies[RoundRobinStrategy] = func() BalancingStrategy {
		return &RoundRobin{
			Current: uint32(0),
		}
	}
}

type RoundRobin struct {
	// Current is the index of the server to be used
	// the next server would be (current + 1) % len(Servers)
	Current uint32
}

func (r *RoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	// normal increament (sl.current + 1) is not thread safe, can lead to race conditions
	// atomic is crucial because multiple goroutines might be requesting the next server simultaneously
	nxt := atomic.AddUint32(&r.Current, uint32(1))

	//TODO: Do not using modulo (nxt % len(sl.Servers)) because it is expensive
	lenS := uint32(len(servers))
	// if nxt >= lenS {
	// 	nxt -= lenS
	// }
	return servers[nxt%lenS], nil
}

// LoadStrategy will try to resolve the strategy name
// default to RoundRobinStrategy if no strategy matched
func LoadStrategy(strategyName string) BalancingStrategy {
	strategy, ok := strategies[strategyName]
	if !ok {
		// Default if provided strategy does not match
		return strategies[RoundRobinStrategy]()
	}
	return strategy()
}
