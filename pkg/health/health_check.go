package health

import (
	"errors"
	"github.com/sachin-404/gobalancer/pkg/domain"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type HealthChecker struct {
	Servers []*domain.Server
	// TODO: take period from configuration
	period int
}

func NewHealthChecker(servers []*domain.Server) (*HealthChecker, error) {
	if len(servers) == 0 {
		return nil, errors.New("expected at least one server, got empty list")
	}
	return &HealthChecker{
		Servers: servers,
		period:  1,
	}, nil
}

func (hc *HealthChecker) Start() {
	log.Info("starting health checker...")
	ticker := time.NewTicker(time.Duration(hc.period) * time.Second)
	for {
		select {
		case <-ticker.C:
			for _, server := range hc.Servers {
				go healthCheck(server)
			}
		}
	}
}

// healthCheck changes the liveness of the server either from live to
// unavailable or the other way around
func healthCheck(server *domain.Server) {
	_, err := net.DialTimeout("tcp", server.URL.Host, time.Second*5)
	if err != nil {
		log.Errorf("error connecting to server %s: %v", server.URL.Host, err)
		old := server.SetLiveness(false)
		if old {
			log.Warnf("health check failed! marking server %s unavailable!", server.URL.Host)
		}
		return
	}
	old := server.SetLiveness(true)
	if !old {
		log.Infof("health check succeeded! making server %s live", server.URL.Host)
	}
}
