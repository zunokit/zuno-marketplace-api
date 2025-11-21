# METADATA SERVICE INTEGRATION SPECIFICATION

**Service**: `zuno-marketplace-metadata`
**Base URL**: `http://localhost:3000/api` (configurable)
**Version**: v1
**Last Verified**: 2025-11-20

---

## 🎯 Overview

Metadata Service provides:
- **Media Upload**: Images, videos, 3D models → ImageKit CDN
- **IPFS Pinning**: Background async pinning to Pinata
- **Metadata Management**: OpenSea-compatible NFT metadata
- **API Key Authentication**: Secure access with scopes
- **Rate Limiting**: Per-key limits via Redis

---

## 📡 API Endpoints

### Media Endpoints

| Method | Path | Auth | Scopes | Description |
|--------|------|------|--------|-------------|
| POST | `/api/media` | Yes | `media:write` | Upload single file |
| POST | `/api/media/batch` | Yes | `media:write` | Upload 1-20 files |
| GET | `/api/media` | Yes | `media:read` | List with pagination |
| GET | `/api/media/[id]` | Yes | `media:read` | Get by ID |
| DELETE | `/api/media/[id]` | Yes | `media:delete` | Delete file |

### Metadata Endpoints

| Method | Path | Auth | Scopes | Description |
|--------|------|------|--------|-------------|
| POST | `/api/metadata` | Yes | `metadata:write` | Create metadata |
| POST | `/api/metadata/batch` | Yes | `metadata:write` | Create 1-50 items |
| GET | `/api/metadata` | Yes | `metadata:read` | List with pagination |
| GET | `/api/metadata/[id]` | Yes | `metadata:read` | Get by ID |
| PUT | `/api/metadata/[id]` | Yes | `metadata:write` | Update metadata |
| DELETE | `/api/metadata/[id]` | Yes | `metadata:delete` | Delete metadata |

### Utility Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | No | Health check |

---

## 🔐 Authentication

### Headers (Multiple Options)

**Option 1: x-api-key** (Recommended)
```
x-api-key: sk_test_abc123...
x-api-version: v1
```

**Option 2: Bearer Token**
```
Authorization: Bearer sk_test_abc123...
x-api-version: v1
```

**Option 3: Session Cookies**
```
Cookie: session=...
x-api-version: v1
```

### API Key Scopes

- `media:read` - Read media files
- `media:write` - Upload media
- `media:delete` - Delete media
- `metadata:read` - Read metadata
- `metadata:write` - Create/update metadata
- `metadata:delete` - Delete metadata

### Rate Limiting

**Response Headers**:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1700000000
```

**429 Error**:
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests",
    "statusCode": 429
  },
  "meta": {
    "retryAfter": 60,
    "requestId": "uuid"
  }
}
```

---

## 📤 Upload Media

### Single File Upload

**Request**:
```bash
curl -X POST http://localhost:3000/api/media \
  -H "x-api-key: sk_test_abc123..." \
  -H "x-api-version: v1" \
  -F "file=@image.png" \
  -F "folder=collections" \
  -F "tags=nft"
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "media_abc123",
    "fileName": "image.png",
    "fileSize": 1048576,
    "mimeType": "image/png",
    "mediaType": "IMAGE",
    "url": "https://ik.imagekit.io/zuno/image.png",
    "thumbnailUrl": null,
    "width": 1920,
    "height": 1080,
    "isPinned": false,
    "ipfsHash": null,
    "ipfsUrl": null,
    "createdAt": "2025-11-20T12:34:56.789Z"
  },
  "meta": {
    "requestId": "uuid",
    "timestamp": "2025-11-20T12:34:56.789Z"
  }
}
```

### Batch Upload (1-20 files)

**Request**:
```bash
curl -X POST http://localhost:3000/api/media/batch \
  -H "x-api-key: sk_test_abc123..." \
  -H "x-api-version: v1" \
  -F "files=@image1.png" \
  -F "files=@image2.jpg" \
  -F "folder=collections"
```

**Response**:
```json
{
  "success": true,
  "data": {
    "success": [
      {
        "id": "media_1",
        "fileName": "image1.png",
        "url": "https://ik.imagekit.io/..."
      }
    ],
    "failed": [],
    "summary": {
      "total": 2,
      "succeeded": 2,
      "failed": 0
    }
  }
}
```

### File Size Limits

```
IMAGE:    50 MB
VIDEO:    100 MB
GIF:      25 MB
MODEL_3D: 100 MB
```

---

## 📝 Create Metadata

### Single Metadata

**Request**:
```bash
curl -X POST http://localhost:3000/api/metadata \
  -H "x-api-key: sk_test_abc123..." \
  -H "x-api-version: v1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My NFT Collection",
    "description": "A cool collection",
    "image": "https://ik.imagekit.io/zuno/logo.png",
    "symbol": "MYNFT",
    "attributes": [
      {
        "traitType": "Rarity",
        "value": "Rare"
      }
    ],
    "creators": [
      {
        "address": "0x123...",
        "verified": false,
        "share": 100
      }
    ],
    "sellerFeeBasisPoints": 500
  }'
```

**Response**:
```json
{
  "success": true,
  "data": {
    "id": "metadata_xyz789",
    "name": "My NFT Collection",
    "image": "https://ik.imagekit.io/zuno/logo.png",
    "version": 1,
    "isLocked": false,
    "isPinned": false,
    "ipfsHash": null,
    "ipfsUrl": null,
    "createdAt": "2025-11-20T12:34:56.789Z"
  },
  "meta": {
    "requestId": "uuid",
    "timestamp": "2025-11-20T12:34:56.789Z"
  }
}
```

### Required Fields

- ✅ `name` (string, 1-100 chars)
- ✅ `image` (valid HTTPS URL)

### Optional Fields

- `description` (string, max 2000 chars)
- `symbol` (string, max 10 chars)
- `bannerImage`, `featuredImage`, `animationUrl`, `externalUrl` (URLs)
- `backgroundColor` (6-char hex without #)
- `attributes` (array, max 50, unique traitTypes)
- `creators` (array, shares must sum to 100)
- `sellerFeeBasisPoints` (0-10000)
- `feeRecipient` (address)

### Batch Metadata (1-50 items)

**Request**:
```bash
curl -X POST http://localhost:3000/api/metadata/batch \
  -H "x-api-key: sk_test_abc123..." \
  -H "x-api-version: v1" \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": [
      {
        "name": "NFT #1",
        "image": "https://example.com/1.png"
      },
      {
        "name": "NFT #2",
        "image": "https://example.com/2.png"
      }
    ]
  }'
```

---

## 🔄 IPFS Pinning Workflow

### How It Works

1. **Upload** → File stored on ImageKit, record in DB (`isPinned: false`)
2. **Background Job** → BullMQ queue processes item
3. **Cron Trigger** → External cron calls `/api/cron/process-media-ipfs`
4. **Pinata Upload** → File pinned to IPFS
5. **Database Update** → `isPinned: true`, `ipfsHash`, `ipfsUrl` set

### Polling for IPFS Status

**Poll this endpoint**:
```bash
curl -X GET http://localhost:3000/api/metadata/metadata_xyz789 \
  -H "x-api-key: sk_test_abc123..." \
  -H "x-api-version: v1"
```

**Check `isPinned` field**:
```json
{
  "success": true,
  "data": {
    "id": "metadata_xyz789",
    "isPinned": true,
    "ipfsHash": "QmXxxx...",
    "ipfsUrl": "https://gateway.pinata.cloud/ipfs/QmXxxx...",
    "pinnedAt": "2025-11-20T12:35:00.000Z"
  }
}
```

### Cron Processing

**Media**: Processes 5 items per run
**Metadata**: Processes 10 items per run

**Typical timing**: 10-30 seconds for pinning

---

## 📊 List & Pagination

### List Metadata with Filters

**Request**:
```bash
curl -X GET "http://localhost:3000/api/metadata?page=1&limit=20&sortBy=createdAt&sortOrder=desc&search=cool&isPinned=true" \
  -H "x-api-key: sk_test_abc123..." \
  -H "x-api-version: v1"
```

**Query Parameters**:
- `page` (default: 1)
- `limit` (default: 20, max: 100)
- `sortBy` (`name`, `createdAt`, `updatedAt`, `version`)
- `sortOrder` (`asc`, `desc`)
- `search` (search in name/description)
- `isPinned` (`true`, `false`)
- `isLocked` (`true`, `false`)

**Response**:
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "metadata_1",
        "name": "Cool NFT",
        "image": "https://...",
        "isPinned": true,
        "createdAt": "2025-11-20T12:34:56.789Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "hasMore": true
    }
  }
}
```

---

## ❌ Error Handling

### Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `UNAUTHORIZED` | 401 | Missing/invalid auth |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid input |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Creator shares must sum to exactly 100",
    "statusCode": 400
  },
  "meta": {
    "requestId": "uuid",
    "timestamp": "2025-11-20T12:34:56.789Z"
  }
}
```

---

## 🔧 Environment Variables

**Backend Integration**:
```env
# Required
METADATA_SERVICE_URL=http://localhost:3000/api
METADATA_SERVICE_API_KEY=sk_test_abc123...

# Optional
METADATA_SERVICE_TIMEOUT=30000
```

**Frontend Integration**:
```env
# Required
NEXT_PUBLIC_METADATA_SERVICE_URL=http://localhost:3000/api
NEXT_PUBLIC_METADATA_SERVICE_API_KEY=sk_public_xyz789...
```

---

## ✅ Integration Checklist

**Backend (Upload Proxy)**:
- [ ] Add `METADATA_SERVICE_URL` to `.env`
- [ ] Add `METADATA_SERVICE_API_KEY` to `.env`
- [ ] Create HTTP client with retry logic
- [ ] Implement multipart/form-data forwarding
- [ ] Handle 429 rate limits with exponential backoff
- [ ] Add timeout handling (30s recommended)

**Frontend (Direct Call)**:
- [ ] Add API key to `.env.local`
- [ ] Implement upload progress tracking
- [ ] Poll for IPFS pinning status
- [ ] Handle file validation (size, type)
- [ ] Display error messages to user
- [ ] Show upload progress UI

---

## 🚨 Common Pitfalls

1. **Forgetting `x-api-version` header** → 400 error for metadata endpoints
2. **Not polling for IPFS status** → baseURI not available
3. **Creator shares not summing to 100** → Validation error
4. **File size exceeds limit** → Upload rejected
5. **Duplicate trait types** → Validation error
6. **Missing rate limit handling** → Service degradation

---

## 📖 Related Documentation

- Backend Upload Proxy: [`backend/04-UPLOAD-PROXY.md`](./backend/04-UPLOAD-PROXY.md)
- Integration Workflows: [`03-INTEGRATION.md`](./03-INTEGRATION.md)
- Main Schema: [`database-schema.md`](../database-schema.md)

---

**Last Updated**: 2025-11-20
**Verified Against**: `zuno-marketplace-metadata` codebase
