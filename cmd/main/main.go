package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/sachin-404/gobalancer/pkg/config"
	"github.com/sachin-404/gobalancer/pkg/domain"
	"github.com/sachin-404/gobalancer/pkg/strategy"

	log "github.com/sirupsen/logrus"
)

var (
	port       = flag.Int("port", 8000, "port to run the load balancer on")
	configFile = flag.String("config", "example/weighted.config.yml", "config file for the load balancer")
)

type LoadBalancer struct {
	// Config is the configuration loaded from the config file
	Config *config.Config
	// ServerList will contain mapping b/w matcher and replicas
	ServerList map[string]*config.ServerList
}

func NewLoadBalancer(cfg *config.Config) *LoadBalancer {
	serverMap := make(map[string]*config.ServerList)
	for _, service := range cfg.Services {
		servers := make([]*domain.Server, 0)
		for _, replica := range service.Replicas {
			u, err := url.Parse(replica.Url)
			if err != nil {
				log.Fatalf("error parsing url: %v", err)
			}
			proxy := httputil.NewSingleHostReverseProxy(u)
			servers = append(servers, &domain.Server{
				URL:      u,
				Proxy:    proxy,
				Metadata: replica.Metadata,
			})
		}
		serverMap[service.Matcher] = &config.ServerList{
			Servers: servers,
			//Current: uint32(0),
			Name:     service.Name,
			Strategy: strategy.LoadStrategy(cfg.Strategy),
		}
	}
	return &LoadBalancer{
		Config:     cfg,
		ServerList: serverMap,
	}
}

// findServiceList looks for the first server list that matches the path (i.e. matcher)
func (l *LoadBalancer) findServiceList(reqPath string) (*config.ServerList, error) {
	//log.Infof("Trying to find the matcher for the request %s", reqPath)
	for matcher, s := range l.ServerList {
		if strings.HasPrefix(reqPath, matcher) {
			//log.Infof("Found service '%s' matching the request", s.Name)
			return s, nil
		}
	}
	return nil, fmt.Errorf("could not find a matcher for %s", reqPath)
}

func (l *LoadBalancer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.Infof("Recieved new request %s", req.Host)
	sl, err := l.findServiceList(req.URL.Path)
	if err != nil {
		log.Error(err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	next, err := sl.Strategy.Next(sl.Servers)
	if err != nil {
		log.Error(err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Infof("Forwarding request to server '%s'", next.URL.Host)
	// forwarding the request to the proxy
	next.Forward(res, req)
}

func main() {
	flag.Parse()
	file, err := os.Open(*configFile)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("error closing file: %v", err)
		}
	}(file)
	if err != nil {
		log.Fatalf("error opening config file: %v", err)
	}
	if file == nil {
		log.Fatalf("config file not provided")
	}
	cfg, err := config.LoadConfig(file)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	lb := NewLoadBalancer(cfg)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: lb,
	}
	log.Infof("starting GoBalancer server with '%s' strategy on port %d", cfg.Strategy, *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
