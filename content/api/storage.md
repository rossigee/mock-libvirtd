---
title: Storage Pools API
---

Storage pools represent disk storage resources.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/storage` | List all storage pools |
| POST | `/api/storage` | Create a storage pool |
| GET | `/api/storage/{id}` | Get storage pool details |
| DELETE | `/api/storage/{id}` | Delete storage pool |

## List Storage Pools

```bash
GET /api/storage
```

Response:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "default-pool",
      "type": "dir",
      "path": "/var/lib/libvirt/images",
      "active": true,
      "capacity": 107374182400
    }
  ],
  "request_id": "uuid"
}
```

## Create Storage Pool

```bash
POST /api/storage
```

Request body:
```json
{
  "name": "string (required)",
  "type": "dir",
  "path": "/var/lib/libvirt/images",
  "capacity": 107374182400
}
```

Defaults:
- `type`: dir
- `path`: /var/lib/libvirt/images
- `capacity`: 100GB (107374182400 bytes)

Response: `201 Created`

### Errors

| Status | Description |
|--------|-------------|
| 400 | Invalid request (missing name) |

## Get Storage Pool

```bash
GET /api/storage/{id}
```

Response: `200 OK`

### Errors

| Status | Description |
|--------|-------------|
| 404 | Storage pool not found |

## Delete Storage Pool

```bash
DELETE /api/storage/{id}
```

Response: `200 OK`

### Errors

| Status | Description |
|--------|-------------|
| 404 | Storage pool not found |

## See Also

- [Volumes API](/api/volumes/) - Storage volumes
- [OpenAPI Spec](/openapi.yaml) - Full specification