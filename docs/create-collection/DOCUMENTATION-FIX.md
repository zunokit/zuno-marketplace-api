# DOCUMENTATION FIX - Upload Proxy Architecture

**Date**: 2025-11-22
**Fixed By**: Claude Code
**Issue**: Mâu thuẫn giữa Frontend Plan và Backend Plan về cách gọi Metadata Service

---

## ❌ VẤN ĐỀ BAN ĐẦU

Tài liệu có **2 luồng mâu thuẫn**:

### Backend Plan (04-UPLOAD-PROXY.md) - ĐÚNG ✅
```
Frontend → Backend Upload Proxy → Metadata Service
         ↓ JWT token           ↓ API key (secret)
```

### Frontend Plan (02-FRONTEND-PLAN.md) - SAI ❌
```
Frontend → Metadata Service (direct call)
         ↓ API key exposed in NEXT_PUBLIC env var
```

**Nguyên nhân**: Tài liệu frontend chưa được cập nhật sau khi thêm Upload Proxy (Phase 4).

---

## ✅ ĐÃ SỬA

### 1. File: `02-FRONTEND-PLAN.md`

**Thay đổi**:
- ❌ Xóa: `MetadataServiceClient` (gọi trực tiếp Metadata Service)
- ✅ Thêm: `UploadProxyClient` (gọi Backend Upload Proxy)
- ❌ Xóa: `NEXT_PUBLIC_METADATA_SERVICE_API_KEY` env var
- ✅ Thêm: JWT authentication từ auth context
- ✅ Thêm: Comments giải thích flow: Frontend → Backend → Metadata

**Code thay đổi**:
```typescript
// ❌ CŨ - SAI
const metadataClient = new MetadataServiceClient(
  process.env.NEXT_PUBLIC_METADATA_SERVICE_URL,
  process.env.NEXT_PUBLIC_METADATA_SERVICE_API_KEY  // Lộ API key!
);

// ✅ MỚI - ĐÚNG
const uploadClient = createUploadProxyClient(getAccessToken);  // JWT token
// Calls: /api/upload/* (Backend proxy)
// Backend forwards với API key (secret)
```

---

### 2. File: `03-INTEGRATION.md`

**Thay đổi**:
- ✅ Cập nhật bảng integration: Frontend ↔ Backend ↔ Metadata
- ✅ Sửa workflow bước 3, 6, 7: Thêm Backend proxy vào flow
- ✅ Cập nhật environment variables
- ✅ Thêm security notes

**Workflow cũ vs mới**:
```
❌ CŨ:
Frontend → POST /api/media (Metadata Service)
         x-api-key: exposed

✅ MỚI:
Frontend → POST /api/upload/media (Backend Proxy)
         Authorization: Bearer JWT
         ↓
Backend → POST /api/media (Metadata Service)
        x-api-key: secret 🔒
```

---

### 3. File: `00-OVERVIEW.md`

**Thay đổi**:
- ✅ Cập nhật bước 2-4 trong E2E flow
- ✅ Thêm security note
- ✅ Cập nhật status và last updated date

---

## 🔒 BẢO MẬT

### Trước khi sửa ❌
```
Frontend .env.local:
NEXT_PUBLIC_METADATA_SERVICE_API_KEY=sk_xyz123...
                                    ↑ Lộ ra browser!
```

### Sau khi sửa ✅
```
Frontend .env.local:
NEXT_PUBLIC_BACKEND_URL=http://localhost:8081
                       ↑ Chỉ có URL backend

Backend .env:
METADATA_SERVICE_API_KEY=sk_xyz123...
                        ↑ Giữ bí mật trên server
```

---

## 📊 FILES UPDATED

1. ✅ `docs/create-collection/02-FRONTEND-PLAN.md`
   - MetadataServiceClient → UploadProxyClient
   - Xóa NEXT_PUBLIC_METADATA_SERVICE_API_KEY
   - Thêm JWT authentication

2. ✅ `docs/create-collection/03-INTEGRATION.md`
   - Cập nhật integration matrix
   - Sửa workflow steps 3, 6, 7
   - Thêm security notes

3. ✅ `docs/create-collection/00-OVERVIEW.md`
   - Cập nhật E2E flow
   - Thêm security note

4. ✅ `docs/create-collection/01-BACKEND-PLAN.md`
   - Không thay đổi (đã đúng từ đầu)

5. ✅ `docs/create-collection/backend/04-UPLOAD-PROXY.md`
   - Không thay đổi (đã đúng từ đầu)

---

## ✅ XÁC NHẬN LUỒNG ĐÚNG

```
┌──────────────┐
│   Frontend   │
│              │
│ JWT Token    │
└──────┬───────┘
       │ POST /api/upload/media
       │ Authorization: Bearer <JWT>
       ▼
┌──────────────────────────┐
│   Backend (Gateway)      │
│                          │
│ 1. Verify JWT ✅         │
│ 2. Validate file ✅      │
│ 3. Forward request       │
└──────┬───────────────────┘
       │ POST http://localhost:3001/api/media
       │ x-api-key: <SECRET_KEY> 🔒
       ▼
┌──────────────────────────┐
│   Metadata Service       │
│                          │
│ 1. Verify API key ✅     │
│ 2. Upload to ImageKit    │
│ 3. Pin to IPFS (async)   │
└──────┬───────────────────┘
       │ Response
       ▼
    Frontend
```

---

## 🎯 KẾT QUẢ

- ✅ Tài liệu nhất quán giữa Frontend và Backend
- ✅ API key được bảo mật (không lộ ở frontend)
- ✅ Luồng upload đúng theo thiết kế Phase 4
- ✅ Security best practices được áp dụng

---

**Verified by**: Claude Code
**Date**: 2025-11-22
