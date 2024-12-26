package main

import (
	"flag"
	"fmt"
	"github.com/sachin-404/gobalancer/pkg/config"
	"net/http"
	"net/http/httputil"
	"net/url"

	log "github.com/sirupsen/logrus"
)

var port = flag.Int("port", 8000, "port to run the load balancer on")

type LoadBalancer struct {
	Config     *config.Config
	ServerList *config.ServerList
}

func NewLoadBalancer(cfg *config.Config) *LoadBalancer {
	servers := make([]*config.Server, 0)
	for _, service := range cfg.Services {
		// TODO: dont ignore the names
		for _, replica := range service.Replicas {
			u, err := url.Parse(replica)
			if err != nil {
				log.Fatalf("error parsing url: %v", err)
			}
			proxy := httputil.NewSingleHostReverseProxy(u)
			servers = append(servers, &config.Server{
				URL:   u,
				Proxy: proxy,
			})
		}
	}
	return &LoadBalancer{
		Config: cfg,
		ServerList: &config.ServerList{
			Servers: servers,
			Current: uint32(0),
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
	l.ServerList.Servers[next].Proxy.ServeHTTP(res, req)
}

func main() {
	flag.Parse()

	cfg := &config.Config{
		Services: []config.Service{
			{
				Name:     "service1",
				Replicas: []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"},
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
