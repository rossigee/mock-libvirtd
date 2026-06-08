# Mock Libvirtd Service

A mock libvirtd service for E2E testing of libvirtd-based applications without requiring actual hypervisor infrastructure. Provides HTTP/REST endpoints for managing virtual machines, networks, and storage pools.

## Overview

This service emulates the libvirtd daemon API for testing purposes. It supports:

- **Domains**: Virtual machine lifecycle management (create, list, get, update, delete)
- **Networks**: Virtual network management (create, list, get, delete)
- **Storage**: Storage pool management (create, list, get, delete)

All data is stored in-memory and reset on service restart.

## Endpoints

### Health Checks

- `GET /health` - Liveness probe (always returns 200)
- `GET /ready` - Readiness probe (returns 200 when ready)

### Domains (Virtual Machines)

- `GET /api/domains` - List all domains
- `POST /api/domains` - Create a new domain
  - Body: `{"name": string, "memory": int, "cpus": int}`
  - Default memory: 512 MB
  - Default CPUs: 1
- `GET /api/domains/{id}` - Get domain details
- `PUT /api/domains/{id}` - Update domain
  - Body: `{"state": string, "memory": int, "cpus": int}`
- `DELETE /api/domains/{id}` - Delete domain

### Networks

- `GET /api/networks` - List all networks
- `POST /api/networks` - Create a new network
  - Body: `{"name": string, "bridge": string}`
  - Default bridge: virbr0
- `GET /api/networks/{id}` - Get network details
- `DELETE /api/networks/{id}` - Delete network

### Storage Pools

- `GET /api/storage` - List all storage pools
- `POST /api/storage` - Create a new storage pool
  - Body: `{"name": string, "type": string, "path": string, "capacity": int}`
  - Default type: dir
  - Default path: /var/lib/libvirt/images
  - Default capacity: 100GB (107374182400 bytes)
- `GET /api/storage/{id}` - Get storage pool details
- `DELETE /api/storage/{id}` - Delete storage pool

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | 8080 | HTTP server port |
| `LOG_LEVEL` | info | Logging level (debug, info, warn, error) |
| `GIN_MODE` | release | Gin framework mode (debug, release) |

## Example Usage

### Local Development

```bash
# Install dependencies
go mod tidy

# Run lint checks
make lint

# Run tests
make test

# Start server
make run

# In another terminal, test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/domains

# Create a domain
curl -X POST http://localhost:8080/api/domains \
  -H "Content-Type: application/json" \
  -d '{"name": "test-vm", "memory": 1024, "cpus": 2}'
```

### Docker

```bash
# Build image
make docker-build

# Run container
make docker-run

# Test in another terminal
curl http://localhost:8080/health
```

### Kubernetes

```bash
# Deploy to cluster
kubectl apply -f k8s/manifest.yaml

# Check status
kubectl get pods -n mock-services
kubectl logs -n mock-services -l app=mock-libvirtd

# Port forward for local testing
kubectl port-forward -n mock-services svc/mock-libvirtd 8080:8080
curl http://localhost:8080/health
```

## Response Format

All responses are JSON. Successful responses follow this format:

```json
{
  "data": [...],
  "request_id": "uuid"
}
```

Error responses:

```json
{
  "error": "description",
  "details": {...},
  "request_id": "uuid"
}
```

## Request Tracing

All requests include a unique `X-Request-ID` header for tracing:

```bash
curl -H "X-Request-ID: my-trace-id" http://localhost:8080/api/domains
```

If not provided, a UUID is generated automatically.

## Logging

All logs are structured JSON with the following fields:

- `timestamp` - RFC3339 timestamp
- `level` - Log level (INFO, WARN, ERROR)
- `msg` - Log message
- `request_id` - Request trace ID
- Additional context fields depending on operation

## Testing

```bash
# Run all tests
go test -v -race ./...

# With coverage report
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Run specific test
go test -v -run TestDomainCreate ./internal/handler
```

## Building

### Local Build

```bash
go build -o mock-libvirtd ./cmd/main
./mock-libvirtd
```

### Docker Build

The Dockerfile uses multi-stage builds:

1. **Builder stage**: Go 1.26.4, downloads dependencies, runs golangci-lint and tests
2. **Final stage**: scratch image with only the binary and CA certificates

Target image size: ≤50MB

```bash
docker build -t mock-libvirtd:latest .
```

## Architecture

```
cmd/main/
  └── main.go              # Entry point, Gin setup, route registration

internal/
  ├── handler/
  │   ├── health.go        # Liveness/readiness probes
  │   ├── domains.go       # Domain management
  │   ├── networks.go      # Network management
  │   ├── storage.go       # Storage pool management
  │   └── handlers_test.go # Handler tests
  └── middleware/
      └── middleware.go    # Gin middleware stack
```

## Standards

This service follows the mock-servers standards:

- **Language**: Go 1.26.4
- **Framework**: Gin for HTTP
- **Logging**: stdlib `log/slog` with JSON output
- **Linting**: golangci-lint 2.12.2 (no exceptions)
- **Testing**: ≥70% coverage on handler logic
- **Kubernetes**: Deployment with health probes and resource limits
- **CI/CD**: GitHub Actions with lint → test → build pipeline

See [STANDARDS.md](../docs/STANDARDS.md) in the mock-servers repo for full guidelines.

## Contributing

Follow the [Contributing Guidelines](../docs/CONTRIBUTING.md) in the mock-servers repo.

## License

Mocks for testing purposes only.
