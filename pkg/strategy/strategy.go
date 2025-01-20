package strategy

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"

	"github.com/sachin-404/gobalancer/pkg/domain"
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
	strategies[WeightedRoundRobinStrategy] = func() BalancingStrategy {
		return &WeightedRoundRobin{mu: sync.Mutex{}}
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

// WeightedRoundRobin is similar to RoundRobin but the diff is it takes compute power
// into consideration.The compute power of a server is given as an integer, it
// represents the fraction of requests that one server can handle over the other
type WeightedRoundRobin struct {
	mu sync.Mutex
	// Note: This is making the assumption that the server list coming through
	// the Next function won't change between successive calls. Changing the server
	// list would cause the strategy to fail, panic or not route properly
	//
	// Count keeps the track of number of requests server 'i' has processed
	Count []int
	// Current is the index of the last server that processed the request
	Current int
}

func (r *WeightedRoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Count == nil {
		// First time using the strategy
		r.Count = make([]int, len(servers))
		r.Current = 0
	}
	capacity := servers[r.Current].GetMetaOrDefaultInt("weight", 1)
	if r.Count[r.Current] <= capacity {
		r.Count[r.Current]++
		return servers[r.Current], nil
	}

	// Server is at its capacity, reset the current one ane move on to the next one
	r.Count[r.Current] = 0
	r.Current = (r.Current + 1) % len(servers)
	return servers[r.Current], nil
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
