## Remote Registry API v1 Documentation

### Base URL
```
http://localhost:8080/remote
```

### Authentication
All endpoints support optional authentication via:
- **JWT Token**: `Authorization: Bearer <token>`
- **API Key**: `X-API-Key: <api-key>`

---

## Registry Endpoints

### Create Registry
**POST** `/registries`

**Request Body:**
```json
{
  "name": "my-registry",
  "description": "My package registry",
  "admin": ["admin@example.com"],
  "customValues": {
    "team": "devops",
    "environment": "production"
  }
}
```

**Response (201 Created):**
```json
{
  "name": "my-registry",
  "description": "My package registry",
  "admin": ["admin@example.com"],
  "customValues": {
    "team": "devops",
    "environment": "production"
  },
  "createdAt": "2025-11-19T10:00:00Z",
  "updatedAt": "2025-11-19T10:00:00Z"
}
```

### List Registries
**GET** `/registries`

**Response (200 OK):**
```json
[
  {
    "name": "my-registry",
    "description": "My package registry",
    "admin": ["admin@example.com"],
    "customValues": {},
    "createdAt": "2025-11-19T10:00:00Z",
    "updatedAt": "2025-11-19T10:00:00Z"
  }
]
```

### Get Registry
**GET** `/registries/:name`

**Response (200 OK):**
```json
{
  "name": "my-registry",
  "description": "My package registry",
  "admin": ["admin@example.com"],
  "customValues": {},
  "createdAt": "2025-11-19T10:00:00Z",
  "updatedAt": "2025-11-19T10:00:00Z"
}
```

### Update Registry
**PUT** `/registries/:name`

**Request Body:**
```json
{
  "description": "Updated description",
  "admin": ["admin@example.com", "admin2@example.com"]
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Registry updated successfully"
}
```

### Delete Registry
**DELETE** `/registries/:name`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Registry deleted successfully"
}
```

### Get CDT-Compatible Index
**GET** `/registries/:name/index.json`

**Response (200 OK):**
```json
[
  {
    "name": "my-package",
    "version": "1.0.0",
    "url": "http://example.com/my-package-1.0.0.zip",
    "checksum": "abc123...",
    "startPartition": 0,
    "endPartition": 9
  }
]
```

---

## Package Endpoints

### Create Package
**POST** `/registries/:registry/packages`

**Request Body:**
```json
{
  "name": "my-package",
  "description": "My awesome package",
  "admin": ["maintainer@example.com"],
  "customValues": {
    "language": "go"
  }
}
```

**Response (201 Created):**
```json
{
  "name": "my-package",
  "description": "My awesome package",
  "admin": ["maintainer@example.com"],
  "customValues": {
    "language": "go"
  },
  "createdAt": "2025-11-19T10:00:00Z",
  "updatedAt": "2025-11-19T10:00:00Z"
}
```

### List Packages
**GET** `/registries/:registry/packages`

**Response (200 OK):**
```json
[
  {
    "name": "my-package",
    "description": "My awesome package",
    "admin": ["maintainer@example.com"],
    "customValues": {},
    "createdAt": "2025-11-19T10:00:00Z",
    "updatedAt": "2025-11-19T10:00:00Z"
  }
]
```

### Get Package
**GET** `/registries/:registry/packages/:package`

### Update Package
**PUT** `/registries/:registry/packages/:package`

### Delete Package
**DELETE** `/registries/:registry/packages/:package`

---

## Version Endpoints

### Publish Version
**POST** `/registries/:registry/packages/:package/versions`

**Request Body:**
```json
{
  "version": "1.0.0",
  "url": "http://example.com/my-package-1.0.0.zip",
  "checksum": "sha256:abc123...",
  "startPartition": 0,
  "endPartition": 9
}
```

**Response (201 Created):**
```json
{
  "version": "1.0.0",
  "url": "http://example.com/my-package-1.0.0.zip",
  "checksum": "sha256:abc123...",
  "startPartition": 0,
  "endPartition": 9,
  "createdAt": "2025-11-19T10:00:00Z",
  "updatedAt": "2025-11-19T10:00:00Z"
}
```

### List Versions
**GET** `/registries/:registry/packages/:package/versions`

**Response (200 OK):**
```json
[
  {
    "version": "1.0.0",
    "url": "http://example.com/my-package-1.0.0.zip",
    "checksum": "sha256:abc123...",
    "startPartition": 0,
    "endPartition": 9,
    "createdAt": "2025-11-19T10:00:00Z",
    "updatedAt": "2025-11-19T10:00:00Z"
  }
]
```

### Get Version
**GET** `/registries/:registry/packages/:package/versions/:version`

### Delete Version
**DELETE** `/registries/:registry/packages/:package/versions/:version`

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

**Common Error Codes:**
- `invalid_request` (400) - Request validation failed
- `unauthorized` (401) - Authentication required or failed
- `not_found` (404) - Resource not found
- `already_exists` (409) - Resource already exists
- `internal_error` (500) - Server error
