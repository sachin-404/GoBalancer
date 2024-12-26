package config

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Service struct {
	Name     string   `yaml:"name"`
	Replicas []string `yaml:"replicas"`
}

// Config is the configuration given to the load balancer
// from a config source
type Config struct {
	Services []Service `yaml:"services"`
	// strategy to be used for load balancing
	Strategy string `yaml:"strategy"`
}

// Server is an instance of a running server
type Server struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy
}

type ServerList struct {
	Servers []*Server

	// current is the index of the server to be used
	// the next server would be (current + 1) % len(Servers)
	Current uint32
}

func (sl *ServerList) Next() uint32 {
	// normal increament (sl.current + 1) is not thread safe, can lead to race conditions
	// atomic is crucial because multiple goroutines might be requesting the next server simultaneously
	nxt := atomic.AddUint32(&sl.Current, uint32(1))

	// not using modulo (nxt % len(sl.Servers)) because it is expensive
	lenS := uint32(len(sl.Servers))
	// if nxt >= lenS {
	// 	nxt -= lenS
	// }
	return nxt % lenS
}
