---
title: Domains API
---

Domains represent virtual machines.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/domains` | List all domains |
| POST | `/api/domains` | Create a domain |
| GET | `/api/domains/{id}` | Get domain details |
| PUT | `/api/domains/{id}` | Update domain |
| DELETE | `/api/domains/{id}` | Delete domain |

## List Domains

```bash
GET /api/domains
```

Response:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "test-vm",
      "state": "running",
      "memory": 1024,
      "cpus": 2,
      "created_at": 1699999999000,
      "started_at": 1699999999000,
      "uptime": 3600,
      "cpu_usage": 25.5,
      "mem_usage": 45.2
    }
  ],
  "request_id": "uuid"
}
```

## Create Domain

```bash
POST /api/domains
```

Request body:
```json
{
  "name": "string (required)",
  "memory": 1024,
  "cpus": 2
}
```

Defaults:
- `memory`: 512 MB
- `cpus`: 1

Response: `201 Created`

### Errors

| Status | Description |
|--------|-------------|
| 400 | Invalid request (missing name, invalid memory/CPU) |
| 503 | Domain limit reached (MAX_DOMAINS) |

```json
{
  "error": "invalid name: cannot be empty",
  "request_id": "uuid"
}
```

## Get Domain

```bash
GET /api/domains/{id}
```

Response: `200 OK`

## Update Domain

```bash
PUT /api/domains/{id}
```

Request body:
```json
{
  "state": "running|paused|shutoff",
  "memory": 2048,
  "cpus": 4
}
```

Only `state` is required for state transitions. Memory and CPU changes are applied immediately.

### State Transitions

Valid states: `running`, `paused`, `shutoff`

Invalid transitions return `409 Conflict`.

### Errors

| Status | Description |
|--------|-------------|
| 404 | Domain not found |
| 409 | Invalid state transition |

```json
{
  "error": "invalid state transition",
  "details": {"from": "shutoff", "to": "paused"},
  "request_id": "uuid"
}
```

## Delete Domain

```bash
DELETE /api/domains/{id}
```

Response: `200 OK`

### Errors

| Status | Description |
|--------|-------------|
| 404 | Domain not found |

```json
{
  "error": "domain not found",
  "request_id": "uuid"
}
```

## See Also

- [State Machine](/state-machine/) - Domain lifecycle
- [OpenAPI Spec](/openapi.yaml) - Full specification