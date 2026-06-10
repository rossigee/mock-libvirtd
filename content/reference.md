---
title: API Reference
---

## Overview

All API endpoints return JSON. Successful responses include a `request_id` for tracing.

## Base URL

```
http://localhost:8080
```

## Response Format

### Success

```json
{
  "data": [...],
  "request_id": "uuid"
}
```

### Error

```json
{
  "error": "description",
  "details": {...},
  "request_id": "uuid"
}
```

Common HTTP status codes:

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (validation error) |
| 404 | Not Found |
| 409 | Conflict (invalid state transition) |
| 429 | Too Many Requests (rate limit) |
| 503 | Service Unavailable (resource limit) |

## Endpoints

### Health Checks

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | HTML home page with live stats |
| GET | `/health` | Liveness probe |
| GET | `/ready` | Readiness probe |
| GET | `/stats` | JSON stats |
| GET | `/metrics` | Prometheus metrics |
| GET | `/loglevel` | Get log level |
| POST | `/loglevel` | Set log level |

### Domains

See [Domains API](/api/domains/)

### Networks

See [Networks API](/api/networks/)

### Storage Pools

See [Storage API](/api/storage/)

### Storage Volumes

See [Volumes API](/api/volumes/)

## OpenAPI Specification

The full OpenAPI specification is available at [/openapi.yaml](/openapi.yaml).

You can visualize it with:
- [Swagger Editor](https://editor.swagger.io/)
- [Redoc](https://redocly.com/redoc/)

## Request Tracing

All requests include a unique `X-Request-ID` header:

```bash
curl -H "X-Request-ID: my-trace-id" http://localhost:8080/api/domains
```

If not provided, a UUID is generated automatically.