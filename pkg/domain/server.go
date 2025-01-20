package domain

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

type Replica struct {
	Url      string            `yaml:"url"`
	Metadata map[string]string `yaml:"metadata"`
}

type Service struct {
	Name string `yaml:"name"`
	// Matcher is a prefix matcher to select service based on the request path
	Matcher string `yaml:"matcher"`
	// Strategy is the loadbalancing strategy used for this service
	Strategy string    `yaml:"strategy"`
	Replicas []Replica `yaml:"replicas"`
}

// Server is an instance of a running server
type Server struct {
	URL      *url.URL
	Proxy    *httputil.ReverseProxy
	Metadata map[string]string
}

func (s *Server) Forward(w http.ResponseWriter, req *http.Request) {
	s.Proxy.ServeHTTP(w, req)
}

// GetMetaOrDefault return the value associated with the given key in the metadata
// or returns the default
func (s *Server) GetMetaOrDefault(key, def string) string {
	v, ok := s.Metadata[key]
	if !ok {
		return def
	}
	return v
}

// GetMetaOrDefaultInt returns the int value associated with the given
// key in the metadata, or returns the default value
func (s *Server) GetMetaOrDefaultInt(key string, def int) int {
	v := s.GetMetaOrDefault(key, strconv.Itoa(def))
	a, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return a
}
