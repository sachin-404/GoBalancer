# GoBalancer

This is a Load balancer written in Go and supports multiple strategies.

The Load balancer takes config input as an `yaml` file. Config consists of services, each service has a name, balancing strategy, a matching pattern, and a list of server replicas.

Sample config files are present in [example](./example) directory.

This is a sample config file:

```yaml
services:
  - name: "UI"
    strategy: RoundRobin
    matcher: /ui
    replicas:
      - url: http://localhost:3001
      - url: http://localhost:3002
  - name: "API"
    matcher: /api/v1
    strategy: WeightedRoundRobin
    replicas:
      - url: http://localhost:8001
        metadata:
          weight: 3
      - url: http://localhost:8002
        metadata:
          weight: 1
```

## Running the Load Balancer

The Load balancer can be run using the following command:

```bash
go run cmd/main/main.go --config example/config.yml
```

Use demo servers to test the load balancer. The demo servers are present in the [cmd/demo/](/cmd/demo/) directory.

```bash
# Start two api servers
go run cmd/demo/api/api_server.go --port=8001
go run cmd/demo/api/api_server.go --port=8002

# Start two ui servers
go run cmd/demo/ui/ui_server.go --port=3001
go run cmd/demo/ui/ui_server.go --port=3002
```

Send requests to the load balancer using `curl http://localhost:8000/ui` or `curl http://localhost:8000/api/v1`. The load balancer will route the requests to the appropriate server based on the strategy.
