---
title: Storage Volumes API
---

Storage volumes represent disk images within a storage pool.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/storage/{pool_id}/volumes` | List volumes in a pool |
| POST | `/api/storage/{pool_id}/volumes` | Create a volume |
| GET | `/api/storage/{pool_id}/volumes/{volume_id}` | Get volume details |
| DELETE | `/api/storage/{pool_id}/volumes/{volume_id}` | Delete a volume |

## List Volumes

```bash
GET /api/storage/{pool_id}/volumes
```

Response:
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "disk.qcow2",
      "pool_id": "pool-uuid",
      "type": "file",
      "size": 10737418240,
      "path": "/var/lib/libvirt/images/disk.qcow2",
      "format": "qcow2"
    }
  ],
  "request_id": "uuid"
}
```

## Create Volume

```bash
POST /api/storage/{pool_id}/volumes
```

Request body:
```json
{
  "name": "string (required)",
  "size": 10737418240,
  "type": "file",
  "format": "qcow2"
}
```

Defaults:
- `type`: file
- `format`: qcow2

Response: `201 Created`

### Errors

| Status | Description |
|--------|-------------|
| 400 | Invalid request (missing name, invalid size) |

### Input Validation

- `name`: Cannot be empty or contain path traversal (`../`)
- `size`: Must be positive

## Get Volume

```bash
GET /api/storage/{pool_id}/volumes/{volume_id}
```

Response: `200 OK`

### Errors

| Status | Description |
|--------|-------------|
| 404 | Volume not found |

## Delete Volume

```bash
DELETE /api/storage/{pool_id}/volumes/{volume_id}
```

Response: `200 OK`

### Errors

| Status | Description |
|--------|-------------|
| 404 | Volume not found |

## See Also

- [Storage Pools API](/api/storage/) - Storage pools
- [OpenAPI Spec](/openapi.yaml) - Full specification