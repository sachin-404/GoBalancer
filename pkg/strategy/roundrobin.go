package strategy

import (
	"errors"
	"github.com/sachin-404/gobalancer/pkg/domain"
	log "github.com/sirupsen/logrus"
	"sync"
)

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
