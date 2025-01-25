package strategy

import (
	"errors"
	"github.com/sachin-404/gobalancer/pkg/domain"
	log "github.com/sirupsen/logrus"
	"sync"
)

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
