---
title: Getting Started
---

## Prerequisites

- Go 1.26+ (for local development)
- Docker (for containerized deployment)
- kubectl (for Kubernetes deployment)

## Quick Start

### Docker

```bash
docker run -d -p 8080:8080 ghcr.io/rossigee/mock-libvirtd
```

### Local Development

```bash
# Clone and setup
git clone https://github.com/rossigee/mock-libvirtd.git
cd mock-libvirtd

# Install dependencies
go mod tidy

# Run the service
go run ./cmd/main
```

### Test the Service

```bash
# Health check
curl http://localhost:8080/health

# List domains (empty)
curl http://localhost:8080/api/domains

# Create a domain
curl -X POST http://localhost:8080/api/domains \
  -H "Content-Type: application/json" \
  -d '{"name": "test-vm", "memory": 1024, "cpus": 2}'

# Get domain details
curl http://localhost:8080/api/domains/{id}

# Start the domain
curl -X PUT http://localhost:8080/api/domains/{id} \
  -H "Content-Type: application/json" \
  -d '{"state": "running"}'

# Stop the domain
curl -X PUT http://localhost:8080/api/domains/{id} \
  -H "Content-Type: application/json" \
  -d '{"state": "shutoff"}'

# Delete the domain
curl -X DELETE http://localhost:8080/api/domains/{id}
```

## Running Tests

```bash
# Run all tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Building

### Binary

```bash
go build -o mock-libvirtd ./cmd/main
./mock-libvirtd
```

### Docker Image

```bash
docker build -t mock-libvirtd:latest .
```

## Next Steps

- [Configuration](/configuration/) - Environment variables and options
- [API Reference](/reference/) - All available endpoints
- [State Machine](/state-machine/) - Understanding domain lifecycle