---
title: Configuration
---

All configuration is via environment variables.

## Server

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | 8080 | HTTP server port |
| `GIN_MODE` | release | Gin framework mode (debug, release) |

## Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | info | Logging level (debug, info, warn, error) |

Runtime log level can also be changed via the `/loglevel` endpoint:

```bash
# Get current level
curl http://localhost:8080/loglevel

# Set new level
curl -X POST http://localhost:8080/loglevel \
  -H "Content-Type: application/json" \
  -d '{"level": "debug"}'
```

## Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_RPS` | 100 | Rate limit requests per second |

## Observability

### OpenTelemetry (OTLP)

| Variable | Default | Description |
|----------|---------|-------------|
| `OTLP_ENDPOINT` | (none) | OpenTelemetry collector endpoint (e.g., `localhost:4317`) |

When set, traces are exported via OTLP/gRPC. When not set, tracing is gracefully disabled.

### Prometheus Metrics

The `/metrics` endpoint exposes Prometheus-compatible metrics:

```
# HELP mock_libvirtd_domains Current number of domains
# TYPE mock_libvirtd_domains gauge
mock_libvirtd_domains 5
```

## Domain State Machine

| Variable | Default | Description |
|----------|---------|-------------|
| `BOOT_TIME_MS` | 1500 | Time for domain to boot (starting → running) in milliseconds |
| `STATE_TICK_RATE_MS` | 100 | How often state machine checks for transitions |

## Resource Limits

| Variable | Default | Description |
|----------|---------|-------------|
| `MAX_DOMAINS` | 1000 | Maximum number of domains |
| `MAX_CPUS` | 256 | Maximum CPUs per domain |
| `MAX_MEMORY_MB` | 1048576 | Maximum memory per domain in MB (default: 1TB) |
| `MIN_CPUS` | 1 | Minimum CPUs per domain |
| `MIN_MEMORY_MB` | 128 | Minimum memory per domain in MB |
| `MAX_NAME_LENGTH` | 255 | Maximum length for resource names |

## Example Usage

### High Limits for Load Testing

```bash
MAX_DOMAINS=10000 RATE_LIMIT_RPS=1000 docker run -d -p 8080:8080 ghcr.io/rossigee/mock-libvirtd
```

### Debug Mode

```bash
LOG_LEVEL=debug go run ./cmd/main
```

### Custom Port

```bash
HTTP_PORT=9090 go run ./cmd/main
```