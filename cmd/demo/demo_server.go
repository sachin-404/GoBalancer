package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var port = flag.Int("port", 8001, "port for demo server")

type DemoServer struct{}

func (ds *DemoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("All good from server %d", *port)))
}

func main() {
	flag.Parse()

	log.Infof("starting demo server on port %d", *port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), &DemoServer{}); err != nil {
		log.Fatalf("error starting demo server: %v", err)
	}
}
