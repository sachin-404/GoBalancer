.PHONY: build lint fmt tidy

# Default config file
config ?= example/config.yml

build:
	go build -o gobalancer cmd/main/main.go

# Run the application (override config with: make run config=custom.yml)
run:
	go run cmd/main/main.go --config $(config)

lint:
	golangci-lint run

fmt:
	go fmt ./...

tidy:
	go mod tidy
	go mod verify