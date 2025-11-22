# FRONTEND IMPLEMENTATION PLAN

**Repository**: `zuno-marketplace-ui`
**Framework**: Next.js 15 (App Router)
**Libraries**: Wagmi v2, RainbowKit, Apollo Client, Zuno SDK
**Package Manager**: pnpm

---

## 🟢 PHASE 1: GraphQL Integration

**Duration Estimate**: Foundation phase
**Dependencies**: Backend Phase 3 (GraphQL schema) complete
**Output**: Type-safe GraphQL hooks

### Task 1.1: GraphQL Schema Definition

**File**: `src/shared/graphql/schemas/collection.graphql`

```graphql
# Enums
enum TokenStandard {
  ERC721
  ERC1155
}

enum CollectionStatus {
  PENDING
  DEPLOYED
  FAILED
  ARCHIVED
}

enum MintStage {
  INACTIVE
  ALLOWLIST
  PUBLIC
}

# Types
type Collection {
  id: ID!
  userId: ID!
  name: String!
  symbol: String!
  description: String
  contractAddress: String
  chainId: String
  tokenStandard: TokenStandard!
  status: CollectionStatus!
  creatorAddress: String!
  deployerAddress: String!
  baseUri: String
  maxSupply: Int
  mintPriceAllowlist: String
  mintPricePublic: String
  mintStartTime: Time
  allowlistStageEnd: Time
  royaltyFeeBps: Int
  royaltyRecipient: String
  metadata: CollectionMetadata
  stats: CollectionStats
  createdAt: Time!
  updatedAt: Time!
  deployedAt: Time
}

type CollectionMetadata {
  imageUrl: String!
  bannerImageUrl: String
  featuredImageUrl: String
  externalUrl: String
  backgroundColor: String
  ipfsHash: String
  ipfsUrl: String
}

type CollectionStats {
  totalItems: Int!
  totalOwners: Int!
  totalSales: Int!
  floorPriceWei: String
  totalVolumeWei: String
  averagePriceWei: String
  lastSaleAt: Time
}

type CollectionConnection {
  items: [Collection!]!
  pageInfo: PageInfo!
}

type PageInfo {
  page: Int!
  totalPages: Int!
  total: Int!
  hasNext: Boolean!
  hasPrev: Boolean!
}

# Inputs
input CreateCollectionInput {
  name: String!
  symbol: String!
  description: String
  tokenStandard: TokenStandard!
  deployerAddress: String!
  metadataUri: String!
  maxSupply: Int
  mintPriceAllowlist: String
  mintPricePublic: String
  mintStartTime: Time
  allowlistStageDurationSeconds: Int
  royaltyFeeBps: Int
  royaltyRecipient: String
  imageUrl: String!
  bannerImageUrl: String
  featuredImageUrl: String
  externalUrl: String
  backgroundColor: String
}

input UpdateCollectionInput {
  contractAddress: String
  chainId: String
  status: CollectionStatus
  deployedAt: Time
  description: String
  externalUrl: String
  bannerImageUrl: String
}

input AddToAllowlistInput {
  collectionId: ID!
  walletAddresses: [String!]!
  maxMintAmount: Int
}

# Queries
extend type Query {
  collection(id: ID!): Collection
  myCollections(page: Int, limit: Int): CollectionConnection!
  collections(
    userId: ID
    status: CollectionStatus
    tokenStandard: TokenStandard
    chainId: String
    page: Int
    limit: Int
    sortBy: String
    sortOrder: String
  ): CollectionConnection!
}

# Mutations
extend type Mutation {
  createCollection(input: CreateCollectionInput!): Collection!
  updateCollection(id: ID!, input: UpdateCollectionInput!): Collection!
  addToAllowlist(input: AddToAllowlistInput!): Boolean!
  deleteCollection(id: ID!): Boolean!
}
```

### Task 1.2: Generate TypeScript Hooks

**Command**:
```bash
cd src/shared/graphql
pnpm codegen
```

**Generated hooks** (in `src/shared/graphql/hooks.ts`):
- `useGetCollectionQuery`
- `useMyCollectionsQuery`
- `useCollectionsQuery`
- `useCreateCollectionMutation`
- `useUpdateCollectionMutation`
- `useAddToAllowlistMutation`
- `useDeleteCollectionMutation`

### Task 1.3: Update TypeScript Types

**File**: `src/shared/types/collection.ts`

```typescript
export type TokenStandard = 'ERC721' | 'ERC1155';
export type CollectionStatus = 'PENDING' | 'DEPLOYED' | 'FAILED' | 'ARCHIVED';
export type MintStage = 'INACTIVE' | 'ALLOWLIST' | 'PUBLIC';

export interface Collection {
  id: string;
  userId: string;
  name: string;
  symbol: string;
  description?: string;
  contractAddress?: string;
  chainId?: string;
  tokenStandard: TokenStandard;
  status: CollectionStatus;
  creatorAddress: string;
  deployerAddress: string;
  baseUri?: string;
  maxSupply?: number;
  mintPriceAllowlist?: string; // Wei
  mintPricePublic?: string; // Wei
  mintStartTime?: Date;
  allowlistStageEnd?: Date;
  royaltyFeeBps?: number;
  royaltyRecipient?: string;
  metadata?: CollectionMetadata;
  stats?: CollectionStats;
  createdAt: Date;
  updatedAt: Date;
  deployedAt?: Date;
}

export interface CollectionMetadata {
  imageUrl: string;
  bannerImageUrl?: string;
  featuredImageUrl?: string;
  externalUrl?: string;
  backgroundColor?: string;
  ipfsHash?: string;
  ipfsUrl?: string;
}

export interface CollectionStats {
  totalItems: number;
  totalOwners: number;
  totalSales: number;
  floorPriceWei?: string;
  totalVolumeWei?: string;
  averagePriceWei?: string;
  lastSaleAt?: Date;
}

export interface CreateCollectionFormData {
  // Step 1: Basic Info
  name: string;
  symbol: string;
  description?: string;
  tokenStandard: TokenStandard;
  category?: string;

  // Step 2: Media
  logoFile?: File;
  bannerFile?: File;
  featuredFile?: File;

  // Step 3: Minting Settings
  maxSupply?: number;
  mintPriceAllowlist?: string; // ETH string
  mintPricePublic?: string; // ETH string
  mintStartDate?: Date;
  allowlistStageDuration?: number; // hours

  // Step 4: Royalty
  royaltyPercentage?: number; // 0-10
  royaltyRecipient?: string;

  // Step 5: Advanced
  externalUrl?: string;
  backgroundColor?: string;

  // Step 6: Allowlist
  allowlistAddresses?: string[];
}

export interface CollectionConnection {
  items: Collection[];
  pageInfo: {
    page: number;
    totalPages: number;
    total: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
}
```

---

## 🟢 PHASE 2: Metadata Service Integration

**Duration Estimate**: API client setup
**Dependencies**: Phase 1 complete
**Output**: Working metadata upload flow

**📖 Full Spec**: See [`METADATA-SERVICE-SPEC.md`](../METADATA-SERVICE-SPEC.md)

**⚠️ IMPORTANT Notes**:
- IPFS pinning is **ASYNC** (~10-30s via background cron)
- Must poll `GET /metadata/{id}` until `isPinned: true`
- Header `x-api-version: v1` is **REQUIRED** for metadata endpoints
- Support 3 auth methods: `x-api-key`, `Authorization: Bearer`, or Session cookies

### Task 2.1: Metadata Service API Client

**File**: `src/shared/api/metadata-client.ts`

```typescript
import axios, { AxiosInstance, AxiosError } from 'axios';

interface MediaResponse {
  id: string;
  fileName: string;
  fileSize: number;
  mimeType: string;
  mediaType: 'IMAGE' | 'VIDEO' | 'GIF' | 'MODEL_3D';
  url: string;
  thumbnailUrl?: string;
  width?: number;
  height?: number;
  isPinned: boolean;              // false initially, true after async pinning
  ipfsHash?: string;              // null until pinned
  ipfsUrl?: string;               // null until pinned, use for baseURI
  pinnedAt?: string;              // timestamp when pinned
  createdAt: string;
}

interface MetadataResponse {
  id: string;
  name: string;
  image: string;
  description?: string;
  symbol?: string;
  bannerImage?: string;
  featuredImage?: string;
  externalUrl?: string;
  backgroundColor?: string;
  isPinned: boolean;
  ipfsHash?: string;
  ipfsUrl?: string;
  pinnedAt?: string;
  createdAt: string;
}

interface CreateMetadataInput {
  name: string;
  image: string;
  description?: string;
  symbol?: string;
  bannerImage?: string;
  featuredImage?: string;
  externalUrl?: string;
  backgroundColor?: string;
}

export class UploadProxyClient {
  private client: AxiosInstance;

  constructor(
    baseURL: string = process.env.NEXT_PUBLIC_BACKEND_URL + '/api/upload',
    getAccessToken: () => string | null
  ) {
    this.client = axios.create({
      baseURL,
      timeout: 30000,
    });

    // Request interceptor to add JWT token
    this.client.interceptors.request.use(
      (config) => {
        const token = getAccessToken();
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (error.response?.status === 401) {
          throw new Error('Unauthorized. Please login again.');
        }
        if (error.response?.status === 429) {
          throw new Error('Rate limit exceeded. Please try again later.');
        }
        throw error;
      }
    );
  }

  async uploadMedia(file: File, options?: { folder?: string; tags?: string[] }): Promise<MediaResponse> {
    const formData = new FormData();
    formData.append('file', file);
    if (options?.folder) {
      formData.append('folder', options.folder);
    }
    if (options?.tags) {
      options.tags.forEach(tag => formData.append('tags', tag));
    }

    // Calls: POST /api/upload/media (Backend proxy)
    // Backend forwards to Metadata Service with API key
    const response = await this.client.post<{ success: boolean; data: MediaResponse }>('/media', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });

    return response.data.data;
  }

  async batchUploadMedia(files: File[], options?: { folder?: string }): Promise<MediaResponse[]> {
    if (files.length > 20) {
      throw new Error('Maximum 20 files per batch upload');
    }

    const formData = new FormData();
    files.forEach(file => formData.append('files', file));
    if (options?.folder) {
      formData.append('folder', options.folder);
    }

    // Calls: POST /api/upload/batch (Backend proxy)
    const response = await this.client.post<{ success: boolean; data: MediaResponse[] }>('/batch', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });

    return response.data.data;
  }

  async createMetadata(input: CreateMetadataInput): Promise<MetadataResponse> {
    // Calls: POST /api/upload/metadata (Backend proxy)
    const response = await this.client.post<{ success: boolean; data: MetadataResponse }>('/metadata', input);
    return response.data.data;
  }

  async getMetadata(id: string): Promise<MetadataResponse> {
    // Calls: GET /api/upload/metadata/:id (Backend proxy)
    const response = await this.client.get<{ success: boolean; data: MetadataResponse }>(`/metadata/${id}`);
    return response.data.data;
  }

  async pollMetadataPinned(id: string, maxRetries: number = 30): Promise<MetadataResponse> {
    let retries = 0;

    while (retries < maxRetries) {
      const metadata = await this.getMetadata(id);

      // Check if IPFS pinning is complete (background cron job)
      if (metadata.isPinned && metadata.ipfsUrl) {
        return metadata;  // Success! Use metadata.ipfsUrl for baseURI
      }

      // Wait 2 seconds before next poll (typical pinning time: 10-30s)
      await new Promise(resolve => setTimeout(resolve, 2000));
      retries++;
    }

    throw new Error('Metadata pinning timeout. Please try again later.');
  }
}

// Factory function that uses auth context
export const createUploadProxyClient = (getAccessToken: () => string | null) => {
  return new UploadProxyClient(
    process.env.NEXT_PUBLIC_BACKEND_URL + '/api/upload',
    getAccessToken
  );
};
```

### Task 2.2: Media Upload Hook

**File**: `src/shared/hooks/useMediaUpload.ts`

```typescript
import { useState, useCallback } from 'react';
import { createUploadProxyClient } from '@/shared/api/upload-proxy-client';
import { useAuth } from '@/shared/hooks/useAuth';  // Your auth context
import { toast } from 'sonner';

interface UseMediaUploadReturn {
  uploadFile: (file: File) => Promise<string>;
  uploadMultiple: (files: File[]) => Promise<string[]>;
  uploading: boolean;
  progress: number;
  error: Error | null;
}

export function useMediaUpload(): UseMediaUploadReturn {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState<Error | null>(null);
  const { getAccessToken } = useAuth();  // Get JWT token from auth context

  const uploadFile = useCallback(async (file: File): Promise<string> => {
    if (!file) {
      throw new Error('No file provided');
    }

    // Validate file size (max 10MB for safety)
    if (file.size > 10 * 1024 * 1024) {
      throw new Error('File too large. Maximum size is 10MB.');
    }

    // Validate file type
    const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg', 'image/gif', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      throw new Error('Invalid file type. Only images are allowed.');
    }

    setUploading(true);
    setProgress(0);
    setError(null);

    try {
      setProgress(30);

      // Create client with JWT token
      const uploadClient = createUploadProxyClient(getAccessToken);

      // Calls Backend Upload Proxy: POST /api/upload/media
      // Backend verifies JWT and forwards to Metadata Service with API key
      const response = await uploadClient.uploadMedia(file, {
        folder: 'collections',
        tags: ['collection-media'],
      });
      setProgress(100);

      return response.url;
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Upload failed');
      setError(error);
      toast.error(error.message);
      throw error;
    } finally {
      setUploading(false);
    }
  }, [getAccessToken]);

  const uploadMultiple = useCallback(async (files: File[]): Promise<string[]> => {
    if (files.length === 0) {
      return [];
    }

    if (files.length > 20) {
      throw new Error('Maximum 20 files per batch');
    }

    setUploading(true);
    setProgress(0);
    setError(null);

    try {
      setProgress(30);

      const uploadClient = createUploadProxyClient(getAccessToken);

      // Calls Backend Upload Proxy: POST /api/upload/batch
      const responses = await uploadClient.batchUploadMedia(files, {
        folder: 'collections',
      });
      setProgress(100);

      return responses.map(r => r.url);
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Batch upload failed');
      setError(error);
      toast.error(error.message);
      throw error;
    } finally {
      setUploading(false);
    }
  }, [getAccessToken]);

  return {
    uploadFile,
    uploadMultiple,
    uploading,
    progress,
    error,
  };
}
```

### Task 2.3: Environment Variables

**File**: `.env.local`

```env
# Backend API (includes GraphQL + Upload Proxy)
NEXT_PUBLIC_BACKEND_URL=http://localhost:8081

# Zuno SDK
NEXT_PUBLIC_ZUNO_API_KEY=your-zuno-api-key
NEXT_PUBLIC_ZUNO_ABI_URL=https://abis.zuno.com/api
```

**Note**:
- ❌ **KHÔNG CẦN** `NEXT_PUBLIC_METADATA_SERVICE_API_KEY` ở frontend
- ✅ API key được giữ bí mật ở backend
- ✅ Frontend gọi `/api/upload/*` với JWT token
- ✅ Backend proxy forward đến Metadata Service

---

## 🟢 PHASE 3: Create Collection Form Enhancement

**Dependencies**: Phase 2 complete
**Output**: Full multi-step form with validation

*See full implementation in separate sections...*

---

## 🟢 PHASE 4: SDK Integration & Transaction Flow

**Dependencies**: Phase 3 complete
**Output**: Working contract deployment

*See full implementation in separate sections...*

---

## 🟢 PHASE 5: Collection Management UI

**Dependencies**: Phase 4 complete
**Output**: My Collections page + Collection detail page

*See full implementation in separate sections...*

---

## 🟢 PHASE 6: Browse Collections UI

**Dependencies**: Phase 5 complete
**Output**: Explore page + Search

*See full implementation in separate sections...*

---

## 🟢 PHASE 7: Testing & Optimization

**Dependencies**: Phase 6 complete
**Output**: Test coverage + Performance optimizations

*See full implementation in separate sections...*

---

**Total Frontend Phases**: 7
**Testing Required**: Unit, Integration, E2E
**Performance Target**: Lighthouse score > 90
