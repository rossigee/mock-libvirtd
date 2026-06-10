---
title: Networks API
---

Networks represent virtual networks.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/networks` | List all networks |
| POST | `/api/networks` | Create a network |
| GET | `/api/networks/{id}` | Get network details |
| DELETE | `/api/networks/{id}` | Delete network |

## List Networks

```bash
GET /api/networks
```

Response:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "test-network",
      "bridge": "virbr0",
      "active": true
    }
  ],
  "request_id": "uuid"
}
```

## Create Network

```bash
POST /api/networks
```

Request body:
```json
{
  "name": "string (required)",
  "bridge": "virbr0"
}
```

Defaults:
- `bridge`: virbr0

Response: `201 Created`

### Errors

| Status | Description |
|--------|-------------|
| 400 | Invalid request (missing name) |

## Get Network

```bash
GET /api/networks/{id}
```

Response: `200 OK`

### Errors

| Status | Description |
|--------|-------------|
| 404 | Network not found |

## Delete Network

```bash
DELETE /api/networks/{id}
```

Response: `200 OK`

### Errors

| Status | Description |
|--------|-------------|
| 404 | Network not found |

## See Also

- [OpenAPI Spec](/openapi.yaml) - Full specification