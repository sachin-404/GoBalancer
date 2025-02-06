package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var UIPort = flag.Int("port", 3000, "port for UI service")

type UIService struct{}

func (u *UIService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf("Hello from UI service %d", *UIPort)))
	if err != nil {
		log.Errorf("Failed to write response: %v", err)
	}
}

func main() {
	flag.Parse()

	log.Infof("Starting UI service on port %d", *UIPort)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *UIPort), &UIService{}); err != nil {
		log.Fatalf("Failed to start UI service: %v", err)
	}
}
