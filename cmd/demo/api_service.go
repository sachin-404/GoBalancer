package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var port = flag.Int("port", 8001, "port for demo service")

type APIService struct{}

func (ds *APIService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf("Hello from API service at %d", *port)))
	if err != nil {
		log.Errorf("Failed to write response: %v", err)
	}
}

func main() {
	flag.Parse()

	log.Infof("starting API service on port %d", *port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), &APIService{}); err != nil {
		log.Fatalf("error starting API service: %v", err)
	}
}
