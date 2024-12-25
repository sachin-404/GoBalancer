package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

var port = flag.Int("port", 8000, "port to run the load balancer on")

type Service struct {
	Name     string
	Replicas []string
}

// Config is the configuration given to the load balancer
// from a config source
type Config struct {
	Services []Service
	// strategy to be used for load balancing
	Strategy string
}

// Server is an instance of a running server
type Server struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

type ServerList struct {
	Servers []*Server

	// current is the index of the server to be used
	// the next server would be (current + 1) % len(Servers)
	current uint32
}

func (sl *ServerList) Next() uint32 {
	// normal increament (sl.current + 1) is not thread safe, can lead to race conditions
	// atomic is crucial because multiple goroutines might be requesting the next server simultaneously
	nxt := atomic.AddUint32(&sl.current, uint32(1))

	// not using modulo (nxt % len(sl.Servers)) because it is expensive
	lenS := uint32(len(sl.Servers))
	// if nxt >= lenS {
	// 	nxt -= lenS
	// }
	return nxt % lenS
}

type LoadBalancer struct {
	Config     *Config
	ServerList *ServerList
}

func NewLoadBalancer(cfg *Config) *LoadBalancer {
	servers := make([]*Server, 0)
	for _, service := range cfg.Services {
		// TODO: dont ignore the names
		for _, replica := range service.Replicas {
			u, err := url.Parse(replica)
			if err != nil {
				log.Fatalf("error parsing url: %v", err)
			}
			proxy := httputil.NewSingleHostReverseProxy(u)
			servers = append(servers, &Server{
				url:   u,
				proxy: proxy,
			})
		}
	}
	return &LoadBalancer{
		Config: cfg,
		ServerList: &ServerList{
			Servers: servers,
			current: uint32(0),
		},
	}
}

func (l *LoadBalancer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// TODO: we need to implement per service forwading, i.e. this method should read the request path
	// say host:port/service1/endpoint, this should be load balanced against service named "service1"
	// and url will be "host{i}:port{i}/endpoint"
	log.Infof("Recieved new request %s", req.Host)
	next := l.ServerList.Next()
	log.Infof("Forwarding request to server number %d", next)
	// forwarding the request to the proxy
	l.ServerList.Servers[next].proxy.ServeHTTP(res, req)
}

func main() {
	flag.Parse()

	cfg := &Config{
		Services: []Service{
			{
				Name:     "service1",
				Replicas: []string{"http://localhost:8001"},
			},
		},
	}

	lb := NewLoadBalancer(cfg)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: lb,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
