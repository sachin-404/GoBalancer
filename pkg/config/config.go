package config

import (
	"github.com/sachin-404/gobalancer/pkg/domain"
	"github.com/sachin-404/gobalancer/pkg/strategy"
)

// Config is the configuration given to the load balancer
// from a config source
type Config struct {
	Services []domain.Service `yaml:"services"`

	// TODO: remove from here
	// strategy to be used for load balancing
	Strategy string `yaml:"strategy"`
}

type ServerList struct {
	// Servers are the replicas
	Servers []*domain.Server
	// Name of the service
	Name string
	// Strategy defines how the server list is load balanced
	// it should never be nil, default is 'RoundRobin'
	Strategy strategy.BalancingStrategy
}
