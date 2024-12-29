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

	log "github.com/sirupsen/logrus"
)

var (
	port       = flag.Int("port", 8000, "port to run the load balancer on")
	configFile = flag.String("config", "example/config.yml", "config file for the load balancer")
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
		servers := make([]*config.Server, 0)
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
		serverMap[service.Matcher] = &config.ServerList{
			Servers: servers,
			//Current: uint32(0),
			Name: service.Name,
		}
	}
	return &LoadBalancer{
		Config:     cfg,
		ServerList: serverMap,
	}
}

// findServiceList looks for the first server list that matches the path (i.e. matcher)
func (l *LoadBalancer) findServiceList(reqPath string) (*config.ServerList, error) {
	log.Infof("Trying to find the matcher for the request %s", reqPath)
	for matcher, s := range l.ServerList {
		if strings.HasPrefix(reqPath, matcher) {
			log.Infof("Found service '%s' matching the request", s.Name)
			return s, nil
		}
	}
	return nil, fmt.Errorf("could not find a matcher for %s", reqPath)
}

func (l *LoadBalancer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.Infof("Recieved new request %s", req.Host)
	sl, err := l.findServiceList(req.URL.Path)
	if err != nil {
		log.Errorf(err.Error())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	next := sl.Next()
	log.Infof("Forwarding request to server number %d", next)
	// forwarding the request to the proxy
	sl.Servers[next].Proxy.ServeHTTP(res, req)
}

func main() {
	flag.Parse()
	file, err := os.Open(*configFile)
	defer file.Close()
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
	log.Infof("starting GoBalncer server on port %d", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
