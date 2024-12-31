package domain

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Service struct {
	Name string `yaml:"name"`
	// Matcher is a prefix matcher to select service based on the request path
	Matcher string `yaml:"matcher"`
	// Strategy is the loadbalancing strategy used for this service
	Strategy string   `yaml:"strategy"`
	Replicas []string `yaml:"replicas"`
}

// Server is an instance of a running server
type Server struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy
}

func (s *Server) Forward(w http.ResponseWriter, req *http.Request) {
	s.Proxy.ServeHTTP(w, req)
}
