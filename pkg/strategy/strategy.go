package strategy

import (
	"errors"
	"github.com/sachin-404/gobalancer/pkg/domain"
	log "github.com/sirupsen/logrus"
	"sync"
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
			mu:      sync.Mutex{},
			Current: 0,
		}
	}

	strategies[WeightedRoundRobinStrategy] = func() BalancingStrategy {
		return &WeightedRoundRobin{mu: sync.Mutex{}}
	}
}

type RoundRobin struct {
	mu sync.Mutex
	// Current is the index of the server to be used
	// the next server would be (current + 1) % len(Servers)
	Current int
}

func (r *RoundRobin) Next(servers []*domain.Server) (*domain.Server, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	seen := 0
	var picked *domain.Server
	for seen < len(servers) {
		picked = servers[r.Current]
		r.Current = (r.Current + 1) % len(servers)
		if picked.IsAlive() {
			break
		}
		seen++
	}

	if picked == nil || seen == len(servers) {
		log.Errorf("no servers available")
		return nil, errors.New("no servers available")
	}

	return picked, nil
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

	seen := 0
	var picked *domain.Server
	for seen < len(servers) {
		picked = servers[r.Current]
		capacity := picked.GetMetaOrDefaultInt("weight", 1)
		if !picked.IsAlive() {
			seen++
			// Current server is not alive so we reset the servers bucket count
			// and try next server in the next loop iteration
			r.Count[r.Current] = 0
			r.Current = (r.Current + 1) % len(servers)
			continue
		}
		if r.Count[r.Current] < capacity {
			r.Count[r.Current]++
			return picked, nil
		}
		// Server is at its capacity, reset the current one and move on to the next one
		r.Count[r.Current] = 0
		r.Current = (r.Current + 1) % len(servers)
	}

	if picked == nil || seen == len(servers) {
		log.Errorf("no servers available")
		return nil, errors.New("no servers available")
	}

	return picked, nil
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
