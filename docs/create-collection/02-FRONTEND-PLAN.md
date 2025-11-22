# FRONTEND IMPLEMENTATION PLAN

**Repository**: `zuno-marketplace-ui`
**Framework**: Next.js 15 (App Router)
**Libraries**: Wagmi v2, RainbowKit, Apollo Client, React Hook Form, Zod
**Package Manager**: pnpm

---

## Current State Analysis

### Existing Form Structure (`CollectionForm.tsx`)

The current form uses **Magic Eden Mint Terminal** style with these fields:

```typescript
interface MintTerminalCreateForm {
  chain: string;              // "sepolia", "base", "polygon"...
  name: string;               // Collection name
  symbol: string;             // Token symbol
  collectionImage?: File;     // Logo image
  artworkMode: "ERC721" | "ERC1155";
  mintStartAt: string;        // ISO date
  description?: string;
  sameArtworkImage?: File;    // ERC1155 artwork
  metadataBaseUrl?: string;   // ERC721 metadata URL
  mintPrice?: string;         // In native token
  royaltyPercent?: number;    // 0-100
  maxSupply?: number | null;
  mintLimitPerWallet?: number | null;
  stages: MintStage[];        // Presale + Public stages
  agreeTos: boolean;
}
```

### Backend GraphQL API Available

```graphql
mutation CreateCollection(input: CreateCollectionInput!): Collection!
mutation UpdateCollection(id: ID!, input: UpdateCollectionInput!): Collection!
mutation AddToAllowlist(input: AddToAllowlistInput!): Boolean!

query collection(id: ID, slug: String, contractAddress: String): Collection
query myCollections(page: Int, limit: Int): CollectionConnection!
query collections(page: Int, limit: Int, sortBy: String, ...): CollectionConnection!

# REST Endpoints
POST /api/upload/media   - Single file upload
POST /api/upload/batch   - Batch file upload
```

---

## PHASE 1: GraphQL Schema & Types

**Output**: Type-safe GraphQL hooks for collections

### Task 1.1: Create Collection GraphQL Schema

**File**: `src/shared/graphql/schemas/collection.graphql`

```graphql
# Enums matching backend
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

enum IndexStatus {
  NOT_INDEXED
  INDEXING
  INDEXED
  FAILED
}

# Types
type Collection {
  id: ID!
  userId: ID!
  name: String!
  symbol: String!
  slug: String
  description: String
  contractAddress: String
  chainId: String!
  tokenStandard: TokenStandard!
  status: CollectionStatus!
  indexStatus: IndexStatus!
  creatorAddress: String!
  deployerAddress: String!

  # Images
  imageUrl: String!
  bannerImageUrl: String
  featuredImageUrl: String

  # Mint settings
  baseUri: String
  maxSupply: Int
  mintPrice: String
  mintPriceAllowlist: String
  mintPricePublic: String
  mintStartTime: String
  allowlistStageEnd: String

  # Royalty
  royaltyFeeBps: Int
  royaltyRecipient: String

  # Relations
  metadata: CollectionMetadata
  stats: CollectionStats

  # Timestamps
  createdAt: String!
  updatedAt: String!
  deployedAt: String
}

type CollectionMetadata {
  websiteUrl: String
  discordUrl: String
  twitterUrl: String
  instagramUrl: String
  mediumUrl: String
  telegramUrl: String
  ipfsHash: String
  ipfsUrl: String
}

type CollectionStats {
  totalItems: Int!
  totalOwners: Int!
  totalVolume: String!
  floorPrice: String!
  totalSales: Int!
  averagePrice: String
  lastSaleAt: String
}

type CollectionConnection {
  items: [Collection!]!
  pageInfo: PageInfo!
}

type PageInfo {
  page: Int!
  limit: Int!
  total: Int!
  totalPages: Int!
  hasNext: Boolean!
  hasPrev: Boolean!
}

# Inputs - mapped from MintTerminalCreateForm
input CreateCollectionInput {
  # Required
  name: String!
  symbol: String!
  tokenStandard: TokenStandard!
  chainId: String!
  deployerAddress: String!
  imageUrl: String!

  # Optional
  description: String
  bannerImageUrl: String
  featuredImageUrl: String

  # Mint settings
  baseUri: String
  maxSupply: Int
  mintPrice: String
  mintPriceAllowlist: String
  mintPricePublic: String
  mintStartTime: String
  allowlistStageDurationSeconds: Int
  mintLimitPerWallet: Int

  # Royalty
  royaltyFeeBps: Int
  royaltyRecipient: String

  # Metadata links
  websiteUrl: String
  discordUrl: String
  twitterUrl: String
}

input UpdateCollectionInput {
  contractAddress: String
  status: CollectionStatus
  deployedAt: String
  description: String
  bannerImageUrl: String
  websiteUrl: String
  discordUrl: String
  twitterUrl: String
}

input AddToAllowlistInput {
  collectionId: ID!
  walletAddresses: [String!]!
  maxMintAmount: Int
}

# Queries
extend type Query {
  collection(id: ID, slug: String, contractAddress: String, chainId: String): Collection
  myCollections(page: Int, limit: Int): CollectionConnection!
  collections(
    page: Int
    limit: Int
    sortBy: String
    sortOrder: String
    category: String
    chainId: String
    isVerified: Boolean
    searchQuery: String
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

### Task 1.2: Add Collection Hooks

**File**: `src/shared/graphql/hooks.ts` (append to existing)

```typescript
import { gql, TypedDocumentNode } from '@apollo/client';
import { useMutation, useQuery, useLazyQuery } from '@apollo/client';

// === COLLECTION OPERATIONS ===

// Create Collection Mutation
const CREATE_COLLECTION: TypedDocumentNode<
  { createCollection: Collection },
  { input: CreateCollectionInput }
> = gql`
  mutation CreateCollection($input: CreateCollectionInput!) {
    createCollection(input: $input) {
      id
      name
      symbol
      slug
      status
      chainId
      imageUrl
      createdAt
    }
  }
`;

export const useCreateCollectionMutation = () => useMutation(CREATE_COLLECTION);

// Update Collection Mutation
const UPDATE_COLLECTION: TypedDocumentNode<
  { updateCollection: Collection },
  { id: string; input: UpdateCollectionInput }
> = gql`
  mutation UpdateCollection($id: ID!, $input: UpdateCollectionInput!) {
    updateCollection(id: $id, input: $input) {
      id
      status
      contractAddress
      deployedAt
      updatedAt
    }
  }
`;

export const useUpdateCollectionMutation = () => useMutation(UPDATE_COLLECTION);

// Add to Allowlist Mutation
const ADD_TO_ALLOWLIST: TypedDocumentNode<
  { addToAllowlist: boolean },
  { input: AddToAllowlistInput }
> = gql`
  mutation AddToAllowlist($input: AddToAllowlistInput!) {
    addToAllowlist(input: $input)
  }
`;

export const useAddToAllowlistMutation = () => useMutation(ADD_TO_ALLOWLIST);

// Delete Collection Mutation
const DELETE_COLLECTION: TypedDocumentNode<
  { deleteCollection: boolean },
  { id: string }
> = gql`
  mutation DeleteCollection($id: ID!) {
    deleteCollection(id: $id)
  }
`;

export const useDeleteCollectionMutation = () => useMutation(DELETE_COLLECTION);

// Get Single Collection Query
const GET_COLLECTION: TypedDocumentNode<
  { collection: Collection | null },
  { id?: string; slug?: string; contractAddress?: string }
> = gql`
  query GetCollection($id: ID, $slug: String, $contractAddress: String) {
    collection(id: $id, slug: $slug, contractAddress: $contractAddress) {
      id
      userId
      name
      symbol
      slug
      description
      contractAddress
      chainId
      tokenStandard
      status
      indexStatus
      creatorAddress
      deployerAddress
      imageUrl
      bannerImageUrl
      featuredImageUrl
      baseUri
      maxSupply
      mintPrice
      mintPriceAllowlist
      mintPricePublic
      mintStartTime
      allowlistStageEnd
      royaltyFeeBps
      royaltyRecipient
      metadata {
        websiteUrl
        discordUrl
        twitterUrl
        instagramUrl
        ipfsHash
        ipfsUrl
      }
      stats {
        totalItems
        totalOwners
        totalVolume
        floorPrice
        totalSales
      }
      createdAt
      updatedAt
      deployedAt
    }
  }
`;

export const useGetCollectionQuery = (variables: { id?: string; slug?: string; contractAddress?: string }) =>
  useQuery(GET_COLLECTION, { variables, skip: !variables.id && !variables.slug && !variables.contractAddress });

export const useGetCollectionLazyQuery = () => useLazyQuery(GET_COLLECTION);

// My Collections Query
const MY_COLLECTIONS: TypedDocumentNode<
  { myCollections: CollectionConnection },
  { page?: number; limit?: number }
> = gql`
  query MyCollections($page: Int, $limit: Int) {
    myCollections(page: $page, limit: $limit) {
      items {
        id
        name
        symbol
        slug
        imageUrl
        chainId
        status
        tokenStandard
        stats {
          totalItems
          totalVolume
          floorPrice
        }
        createdAt
      }
      pageInfo {
        page
        limit
        total
        totalPages
        hasNext
        hasPrev
      }
    }
  }
`;

export const useMyCollectionsQuery = (variables?: { page?: number; limit?: number }) =>
  useQuery(MY_COLLECTIONS, { variables });

// List Collections Query (public browse)
const LIST_COLLECTIONS: TypedDocumentNode<
  { collections: CollectionConnection },
  {
    page?: number;
    limit?: number;
    sortBy?: string;
    sortOrder?: string;
    category?: string;
    chainId?: string;
    isVerified?: boolean;
    searchQuery?: string;
  }
> = gql`
  query ListCollections(
    $page: Int
    $limit: Int
    $sortBy: String
    $sortOrder: String
    $category: String
    $chainId: String
    $isVerified: Boolean
    $searchQuery: String
  ) {
    collections(
      page: $page
      limit: $limit
      sortBy: $sortBy
      sortOrder: $sortOrder
      category: $category
      chainId: $chainId
      isVerified: $isVerified
      searchQuery: $searchQuery
    ) {
      items {
        id
        name
        symbol
        slug
        imageUrl
        bannerImageUrl
        chainId
        status
        tokenStandard
        stats {
          totalItems
          totalOwners
          totalVolume
          floorPrice
          totalSales
        }
        createdAt
      }
      pageInfo {
        page
        limit
        total
        totalPages
        hasNext
        hasPrev
      }
    }
  }
`;

export const useListCollectionsQuery = (variables?: {
  page?: number;
  limit?: number;
  sortBy?: string;
  sortOrder?: string;
  category?: string;
  chainId?: string;
  isVerified?: boolean;
  searchQuery?: string;
}) => useQuery(LIST_COLLECTIONS, { variables });
```

### Task 1.3: TypeScript Types

**File**: `src/shared/types/collection.ts`

```typescript
export type TokenStandard = 'ERC721' | 'ERC1155';
export type CollectionStatus = 'PENDING' | 'DEPLOYED' | 'FAILED' | 'ARCHIVED';
export type IndexStatus = 'NOT_INDEXED' | 'INDEXING' | 'INDEXED' | 'FAILED';

export interface Collection {
  id: string;
  userId: string;
  name: string;
  symbol: string;
  slug?: string;
  description?: string;
  contractAddress?: string;
  chainId: string;
  tokenStandard: TokenStandard;
  status: CollectionStatus;
  indexStatus: IndexStatus;
  creatorAddress: string;
  deployerAddress: string;
  imageUrl: string;
  bannerImageUrl?: string;
  featuredImageUrl?: string;
  baseUri?: string;
  maxSupply?: number;
  mintPrice?: string;
  mintPriceAllowlist?: string;
  mintPricePublic?: string;
  mintStartTime?: string;
  allowlistStageEnd?: string;
  royaltyFeeBps?: number;
  royaltyRecipient?: string;
  metadata?: CollectionMetadata;
  stats?: CollectionStats;
  createdAt: string;
  updatedAt: string;
  deployedAt?: string;
}

export interface CollectionMetadata {
  websiteUrl?: string;
  discordUrl?: string;
  twitterUrl?: string;
  instagramUrl?: string;
  mediumUrl?: string;
  telegramUrl?: string;
  ipfsHash?: string;
  ipfsUrl?: string;
}

export interface CollectionStats {
  totalItems: number;
  totalOwners: number;
  totalVolume: string;
  floorPrice: string;
  totalSales: number;
  averagePrice?: string;
  lastSaleAt?: string;
}

export interface CollectionConnection {
  items: Collection[];
  pageInfo: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
}

export interface CreateCollectionInput {
  name: string;
  symbol: string;
  tokenStandard: TokenStandard;
  chainId: string;
  deployerAddress: string;
  imageUrl: string;
  description?: string;
  bannerImageUrl?: string;
  featuredImageUrl?: string;
  baseUri?: string;
  maxSupply?: number;
  mintPrice?: string;
  mintPriceAllowlist?: string;
  mintPricePublic?: string;
  mintStartTime?: string;
  allowlistStageDurationSeconds?: number;
  mintLimitPerWallet?: number;
  royaltyFeeBps?: number;
  royaltyRecipient?: string;
  websiteUrl?: string;
  discordUrl?: string;
  twitterUrl?: string;
}

export interface UpdateCollectionInput {
  contractAddress?: string;
  status?: CollectionStatus;
  deployedAt?: string;
  description?: string;
  bannerImageUrl?: string;
  websiteUrl?: string;
  discordUrl?: string;
  twitterUrl?: string;
}

export interface AddToAllowlistInput {
  collectionId: string;
  walletAddresses: string[];
  maxMintAmount?: number;
}
```

---

## PHASE 2: Media Upload Service

**Output**: Working upload flow via Backend proxy

### Task 2.1: Upload API Client

**File**: `src/shared/api/upload-client.ts`

```typescript
import axios, { AxiosInstance } from 'axios';
import { graphqlClient } from '@/shared/lib/apollo/apollo-wrapper';

interface MediaUploadResponse {
  success: boolean;
  data: {
    id: string;
    fileName: string;
    fileSize: number;
    mimeType: string;
    url: string;          // ImageKit URL for immediate use
    thumbnailUrl?: string;
    width?: number;
    height?: number;
    ipfsHash?: string;    // Available after async pinning
    ipfsUrl?: string;
  };
}

export class UploadClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8081',
      timeout: 60000, // 60s for large files
    });

    // Add JWT token to requests
    this.client.interceptors.request.use((config) => {
      const token = graphqlClient.getAccessToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });
  }

  /**
   * Upload single media file
   * Returns ImageKit URL immediately (IPFS pinning is async)
   */
  async uploadMedia(
    file: File,
    options?: { folder?: string; tags?: string[] }
  ): Promise<MediaUploadResponse['data']> {
    const formData = new FormData();
    formData.append('file', file);

    if (options?.folder) {
      formData.append('folder', options.folder);
    }
    if (options?.tags) {
      options.tags.forEach(tag => formData.append('tags', tag));
    }

    const response = await this.client.post<MediaUploadResponse>(
      '/api/upload/media',
      formData,
      { headers: { 'Content-Type': 'multipart/form-data' } }
    );

    return response.data.data;
  }

  /**
   * Upload multiple files (max 20)
   */
  async uploadBatch(
    files: File[],
    options?: { folder?: string }
  ): Promise<MediaUploadResponse['data'][]> {
    if (files.length > 20) {
      throw new Error('Maximum 20 files per batch');
    }

    const formData = new FormData();
    files.forEach(file => formData.append('files', file));

    if (options?.folder) {
      formData.append('folder', options.folder);
    }

    const response = await this.client.post<{ success: boolean; data: MediaUploadResponse['data'][] }>(
      '/api/upload/batch',
      formData,
      { headers: { 'Content-Type': 'multipart/form-data' } }
    );

    return response.data.data;
  }
}

export const uploadClient = new UploadClient();
```

### Task 2.2: Media Upload Hook

**File**: `src/shared/hooks/useMediaUpload.ts`

```typescript
import { useState, useCallback } from 'react';
import { uploadClient } from '@/shared/api/upload-client';
import { toast } from 'sonner';

interface UploadState {
  uploading: boolean;
  progress: number;
  error: Error | null;
}

export function useMediaUpload() {
  const [state, setState] = useState<UploadState>({
    uploading: false,
    progress: 0,
    error: null,
  });

  const validateFile = useCallback((file: File) => {
    // Max 10MB
    if (file.size > 10 * 1024 * 1024) {
      throw new Error('File too large. Maximum size is 10MB.');
    }

    // Allowed types
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      throw new Error('Invalid file type. Allowed: JPG, PNG, GIF, WebP.');
    }
  }, []);

  const uploadFile = useCallback(async (file: File): Promise<string> => {
    try {
      validateFile(file);

      setState({ uploading: true, progress: 10, error: null });

      const result = await uploadClient.uploadMedia(file, {
        folder: 'collections',
        tags: ['collection-media'],
      });

      setState({ uploading: false, progress: 100, error: null });

      return result.url; // Return ImageKit URL
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Upload failed');
      setState({ uploading: false, progress: 0, error });
      toast.error(error.message);
      throw error;
    }
  }, [validateFile]);

  const uploadMultiple = useCallback(async (files: File[]): Promise<string[]> => {
    try {
      files.forEach(validateFile);

      setState({ uploading: true, progress: 10, error: null });

      const results = await uploadClient.uploadBatch(files, {
        folder: 'collections',
      });

      setState({ uploading: false, progress: 100, error: null });

      return results.map(r => r.url);
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Batch upload failed');
      setState({ uploading: false, progress: 0, error });
      toast.error(error.message);
      throw error;
    }
  }, [validateFile]);

  const reset = useCallback(() => {
    setState({ uploading: false, progress: 0, error: null });
  }, []);

  return {
    ...state,
    uploadFile,
    uploadMultiple,
    reset,
  };
}
```

---

## PHASE 3: Form Integration

**Output**: CollectionForm connected to backend

### Task 3.1: Create Collection Hook

**File**: `src/modules/mint/create-form/hooks/useCreateCollection.ts`

```typescript
import { useState, useCallback } from 'react';
import { useAccount } from 'wagmi';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { useAuth } from '@/shared/hooks/useAuth';
import { useMediaUpload } from '@/shared/hooks/useMediaUpload';
import {
  useCreateCollectionMutation,
  useAddToAllowlistMutation
} from '@/shared/graphql/hooks';
import { MintTerminalCreateForm } from '@/shared/types/mint';
import { CreateCollectionInput, TokenStandard } from '@/shared/types/collection';
import { parseEther } from 'viem';

export type StepStatus = 'pending' | 'loading' | 'success' | 'error';

interface CreateCollectionState {
  step1Status: StepStatus; // Upload media
  step2Status: StepStatus; // Create collection
  step3Status: StepStatus; // Add allowlist (optional)
  collectionId: string | null;
  error: string | null;
}

// Map chain name to chainId (eip155 format)
const CHAIN_ID_MAP: Record<string, string> = {
  'sepolia': 'eip155:11155111',
  'base': 'eip155:8453',
  'polygon': 'eip155:137',
  'arbitrum': 'eip155:42161',
  'bsc': 'eip155:56',
  'anvil': 'eip155:31337',
};

export function useCreateCollection() {
  const router = useRouter();
  const { address } = useAccount();
  const { isAuthenticated } = useAuth();
  const { uploadFile } = useMediaUpload();
  const [createCollection] = useCreateCollectionMutation();
  const [addToAllowlist] = useAddToAllowlistMutation();

  const [state, setState] = useState<CreateCollectionState>({
    step1Status: 'pending',
    step2Status: 'pending',
    step3Status: 'pending',
    collectionId: null,
    error: null,
  });

  const submit = useCallback(async (formData: MintTerminalCreateForm) => {
    // Validation
    if (!isAuthenticated) {
      toast.error('Please connect your wallet and sign in');
      return;
    }

    if (!address) {
      toast.error('Wallet not connected');
      return;
    }

    setState({
      step1Status: 'loading',
      step2Status: 'pending',
      step3Status: 'pending',
      collectionId: null,
      error: null,
    });

    try {
      // === STEP 1: Upload Media ===
      let imageUrl: string | undefined;
      let artworkUrl: string | undefined;

      // Upload collection image if provided
      if (formData.collectionImage) {
        imageUrl = await uploadFile(formData.collectionImage);
      }

      // Upload artwork image for ERC1155
      if (formData.artworkMode === 'ERC1155' && formData.sameArtworkImage) {
        artworkUrl = await uploadFile(formData.sameArtworkImage);
      }

      setState(prev => ({ ...prev, step1Status: 'success', step2Status: 'loading' }));

      // === STEP 2: Create Collection ===
      // Map form data to GraphQL input
      const chainId = CHAIN_ID_MAP[formData.chain] || `eip155:${formData.chain}`;

      // Calculate allowlist duration in seconds
      const stage = formData.stages[0];
      let allowlistDurationSeconds: number | undefined;
      if (stage.presale?.duration) {
        const { days, hours } = stage.presale.duration;
        allowlistDurationSeconds = (days * 24 * 60 * 60) + (hours * 60 * 60);
      }

      // Build input
      const input: CreateCollectionInput = {
        name: formData.name,
        symbol: formData.symbol,
        tokenStandard: formData.artworkMode as TokenStandard,
        chainId,
        deployerAddress: address,
        imageUrl: imageUrl || artworkUrl || '', // Use artwork as fallback
        description: formData.description,

        // For ERC721, use metadataBaseUrl as baseUri
        // For ERC1155, artwork URL will be in metadata
        baseUri: formData.artworkMode === 'ERC721'
          ? formData.metadataBaseUrl
          : artworkUrl,

        // Supply & limits
        maxSupply: formData.maxSupply || undefined,
        mintLimitPerWallet: formData.mintLimitPerWallet || undefined,

        // Pricing (convert to Wei)
        mintPrice: formData.mintPrice
          ? parseEther(formData.mintPrice).toString()
          : undefined,
        mintPriceAllowlist: stage.presale?.price
          ? parseEther(stage.presale.price).toString()
          : undefined,
        mintPricePublic: stage.public.price
          ? parseEther(stage.public.price).toString()
          : undefined,

        // Timing
        mintStartTime: formData.mintStartAt,
        allowlistStageDurationSeconds: allowlistDurationSeconds,

        // Royalty (convert percentage to basis points)
        royaltyFeeBps: formData.royaltyPercent
          ? formData.royaltyPercent * 100
          : undefined,
        royaltyRecipient: address, // Default to deployer
      };

      const { data, errors } = await createCollection({
        variables: { input },
      });

      if (errors || !data?.createCollection) {
        throw new Error(errors?.[0]?.message || 'Failed to create collection');
      }

      const collectionId = data.createCollection.id;
      setState(prev => ({
        ...prev,
        step2Status: 'success',
        collectionId,
      }));

      // === STEP 3: Add Allowlist (if presale configured) ===
      if (stage.presale?.allowlistAddresses?.length) {
        setState(prev => ({ ...prev, step3Status: 'loading' }));

        await addToAllowlist({
          variables: {
            input: {
              collectionId,
              walletAddresses: stage.presale.allowlistAddresses,
              maxMintAmount: formData.mintLimitPerWallet || undefined,
            },
          },
        });

        setState(prev => ({ ...prev, step3Status: 'success' }));
      } else {
        setState(prev => ({ ...prev, step3Status: 'success' }));
      }

      // Success!
      toast.success('Collection created successfully!');

      // Redirect to collection page or my collections
      router.push(`/collections/${collectionId}`);

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create collection';

      setState(prev => ({
        ...prev,
        step1Status: prev.step1Status === 'loading' ? 'error' : prev.step1Status,
        step2Status: prev.step2Status === 'loading' ? 'error' : prev.step2Status,
        step3Status: prev.step3Status === 'loading' ? 'error' : prev.step3Status,
        error: errorMessage,
      }));

      toast.error(errorMessage);
    }
  }, [isAuthenticated, address, uploadFile, createCollection, addToAllowlist, router]);

  const reset = useCallback(() => {
    setState({
      step1Status: 'pending',
      step2Status: 'pending',
      step3Status: 'pending',
      collectionId: null,
      error: null,
    });
  }, []);

  return {
    ...state,
    submit,
    reset,
    isProcessing: state.step1Status === 'loading' ||
                  state.step2Status === 'loading' ||
                  state.step3Status === 'loading',
  };
}
```

### Task 3.2: Update CollectionForm.tsx

**File**: `src/modules/mint/create-form/components/CollectionForm.tsx`

```typescript
"use client";

import { Button } from "@/shared/components/ui/button";
import { Form } from "@/shared/components/ui/form";
import { RotateCcw } from "lucide-react";
import { CollectionDetails } from "@/modules/mint/create-form/components/CollectionDetails";
import { ArtSection } from "@/modules/mint/create-form/components/ArtSection";
import { MintDetails } from "@/modules/mint/create-form/components/MintDetails";
import { CollectionProcess } from "@/modules/mint/create-form/components/CollectionProcess";
import { useForm } from "react-hook-form";
import { MintTerminalCreateForm, MintTerminalCreateFormSchema } from "@/shared/types/mint";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAuth } from "@/shared/hooks/useAuth";
import { useCreateCollection } from "../hooks/useCreateCollection";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export default function CollectionForm() {
  const { isAuthenticated, isWalletConnected } = useAuth();
  const {
    step1Status,
    step2Status,
    step3Status,
    submit,
    reset: resetProcess,
    isProcessing,
    error
  } = useCreateCollection();

  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const form = useForm<MintTerminalCreateForm>({
    resolver: zodResolver(MintTerminalCreateFormSchema),
    mode: "onChange",
    defaultValues: {
      chain: "sepolia",
      name: "",
      symbol: "",
      collectionImage: undefined,
      artworkMode: "ERC721",
      mintStartAt: new Date(Date.now()).toISOString(),
      description: "",
      sameArtworkImage: undefined,
      metadataBaseUrl: "",
      mintPrice: "0",
      royaltyPercent: 0,
      maxSupply: null,
      mintLimitPerWallet: null,
      stages: [
        {
          public: {
            price: "0",
            duration: null,
          },
        },
      ],
      agreeTos: true,
    },
  });

  const handleClearForm = () => {
    form.reset();
    resetProcess();
  };

  const onSubmit = async (data: MintTerminalCreateForm) => {
    // Validate authentication
    if (!isWalletConnected) {
      toast.error("Please connect your wallet first");
      return;
    }

    if (!isAuthenticated) {
      toast.error("Please sign in with your wallet");
      return;
    }

    // Validate form
    const { success, error } = MintTerminalCreateFormSchema.safeParse(data);
    if (!success) {
      console.error("Validation error:", error);
      toast.error("Please fix form errors before submitting");
      return;
    }

    // Open progress dialog and submit
    setIsDialogOpen(true);
    await submit(data);
  };

  // Close dialog on error after 3s
  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => {
        setIsDialogOpen(false);
      }, 3000);
      return () => clearTimeout(timer);
    }
  }, [error]);

  return (
    <div className="max-w-3xl mx-auto">
      <div className="flex justify-end mb-2">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleClearForm}
          className="text-muted-foreground hover:text-foreground dark:hover:text-white"
        >
          <RotateCcw className="h-4 w-4 mr-1" /> Clear Form
        </Button>
      </div>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <CollectionDetails />
          <ArtSection />
          <MintDetails />

          {/* Terms of Service */}
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="agreeTos"
              {...form.register("agreeTos")}
              className="rounded border-border"
            />
            <label htmlFor="agreeTos" className="text-sm text-foreground dark:text-white">
              I agree to the Terms of Service
            </label>
          </div>
          {form.formState.errors.agreeTos && (
            <p className="text-destructive text-sm">{form.formState.errors.agreeTos.message}</p>
          )}

          <Button
            type="submit"
            className="w-full bg-primary hover:bg-primary/90 text-primary-foreground"
            disabled={
              form.formState.isSubmitting ||
              !form.formState.isValid ||
              isProcessing ||
              !isWalletConnected
            }
          >
            {!isWalletConnected
              ? "Connect Wallet"
              : !isAuthenticated
                ? "Sign In Required"
                : isProcessing
                  ? "Creating..."
                  : "Create Collection"
            }
          </Button>
        </form>
      </Form>

      <CollectionProcess
        isOpen={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        step1Status={step1Status}
        step2Status={step2Status}
        step3Status={step3Status}
      />
    </div>
  );
}
```

### Task 3.3: Update CollectionProcess Dialog

**File**: `src/modules/mint/create-form/components/CollectionProcess.tsx`

```typescript
"use client";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/shared/components/ui/dialog";
import { CheckCircle2, Loader2, XCircle, Circle } from "lucide-react";
import { StepStatus } from "../hooks/useCreateCollection";

interface CollectionProcessProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  step1Status: StepStatus;
  step2Status: StepStatus;
  step3Status?: StepStatus;
}

function StepIcon({ status }: { status: StepStatus }) {
  switch (status) {
    case 'loading':
      return <Loader2 className="h-5 w-5 animate-spin text-primary" />;
    case 'success':
      return <CheckCircle2 className="h-5 w-5 text-success" />;
    case 'error':
      return <XCircle className="h-5 w-5 text-destructive" />;
    default:
      return <Circle className="h-5 w-5 text-muted-foreground" />;
  }
}

export function CollectionProcess({
  isOpen,
  onOpenChange,
  step1Status,
  step2Status,
  step3Status = 'pending',
}: CollectionProcessProps) {
  const steps = [
    {
      title: "Uploading Media",
      description: "Uploading collection images to storage",
      status: step1Status,
    },
    {
      title: "Creating Collection",
      description: "Saving collection data to database",
      status: step2Status,
    },
    {
      title: "Configuring Allowlist",
      description: "Setting up presale allowlist addresses",
      status: step3Status,
    },
  ];

  const allDone = step1Status === 'success' &&
                  step2Status === 'success' &&
                  step3Status === 'success';

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {allDone ? "Collection Created!" : "Creating Collection..."}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {steps.map((step, index) => (
            <div
              key={index}
              className={`flex items-start gap-3 p-3 rounded-lg ${
                step.status === 'loading'
                  ? 'bg-primary/10'
                  : step.status === 'error'
                    ? 'bg-destructive/10'
                    : step.status === 'success'
                      ? 'bg-success/10'
                      : 'bg-muted/50'
              }`}
            >
              <StepIcon status={step.status} />
              <div className="flex-1">
                <p className="font-medium text-sm">{step.title}</p>
                <p className="text-xs text-muted-foreground">{step.description}</p>
              </div>
            </div>
          ))}
        </div>

        {allDone && (
          <p className="text-sm text-center text-muted-foreground">
            Redirecting to your collection...
          </p>
        )}
      </DialogContent>
    </Dialog>
  );
}
```

---

## PHASE 4: Environment Configuration

### Task 4.1: Environment Variables

**File**: `.env.local`

```env
# Backend API (GraphQL + Upload endpoints)
NEXT_PUBLIC_BACKEND_URL=http://localhost:8081

# GraphQL endpoint
NEXT_PUBLIC_GRAPHQL_URL=http://localhost:8081/graphql

# Optional: Zuno SDK for contract deployment
NEXT_PUBLIC_ZUNO_API_KEY=your-zuno-api-key
```

**Note**:
- API key cho Metadata Service được giữ ở backend
- Frontend chỉ cần JWT token để authenticate

---

## PHASE 5: My Collections Page

**Output**: Display user's collections

### Task 5.1: My Collections Page

**File**: `src/app/(user)/my-collections/page.tsx`

```typescript
"use client";

import { useMyCollectionsQuery } from "@/shared/graphql/hooks";
import { useAuth } from "@/shared/hooks/useAuth";
import { CollectionCard } from "@/modules/collections/components/CollectionCard";
import { Button } from "@/shared/components/ui/button";
import { Skeleton } from "@/shared/components/ui/skeleton";
import { Plus } from "lucide-react";
import Link from "next/link";

export default function MyCollectionsPage() {
  const { isAuthenticated } = useAuth();
  const { data, loading, error } = useMyCollectionsQuery({
    page: 1,
    limit: 20
  });

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto py-8 text-center">
        <h1 className="text-2xl font-bold mb-4">My Collections</h1>
        <p className="text-muted-foreground mb-4">
          Please connect your wallet and sign in to view your collections.
        </p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold">My Collections</h1>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(6)].map((_, i) => (
            <Skeleton key={i} className="h-64 rounded-lg" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto py-8 text-center">
        <h1 className="text-2xl font-bold mb-4">My Collections</h1>
        <p className="text-destructive">Error: {error.message}</p>
      </div>
    );
  }

  const collections = data?.myCollections.items || [];

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">My Collections</h1>
        <Link href="/create">
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Create Collection
          </Button>
        </Link>
      </div>

      {collections.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground mb-4">
            You haven't created any collections yet.
          </p>
          <Link href="/create">
            <Button>Create Your First Collection</Button>
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {collections.map((collection) => (
            <CollectionCard key={collection.id} collection={collection} />
          ))}
        </div>
      )}

      {/* Pagination */}
      {data?.myCollections.pageInfo && data.myCollections.pageInfo.totalPages > 1 && (
        <div className="flex justify-center mt-8">
          {/* Add pagination component */}
        </div>
      )}
    </div>
  );
}
```

---

## Summary

### Files to Create/Modify

**New Files**:
1. `src/shared/graphql/schemas/collection.graphql` - GraphQL schema
2. `src/shared/types/collection.ts` - TypeScript types
3. `src/shared/api/upload-client.ts` - Upload API client
4. `src/shared/hooks/useMediaUpload.ts` - Upload hook
5. `src/modules/mint/create-form/hooks/useCreateCollection.ts` - Create collection hook
6. `src/app/(user)/my-collections/page.tsx` - My collections page

**Modified Files**:
1. `src/shared/graphql/hooks.ts` - Add collection hooks
2. `src/modules/mint/create-form/components/CollectionForm.tsx` - Integrate with backend
3. `src/modules/mint/create-form/components/CollectionProcess.tsx` - Real progress status

### Field Mapping: Form → Backend

| Form Field | GraphQL Input | Notes |
|------------|---------------|-------|
| chain | chainId | Convert to eip155 format |
| name | name | Direct |
| symbol | symbol | Direct |
| collectionImage | imageUrl | Upload first, get URL |
| artworkMode | tokenStandard | ERC721/ERC1155 |
| description | description | Direct |
| sameArtworkImage | baseUri (ERC1155) | Upload first |
| metadataBaseUrl | baseUri (ERC721) | Direct |
| mintPrice | mintPrice | Convert to Wei |
| royaltyPercent | royaltyFeeBps | × 100 |
| maxSupply | maxSupply | Direct |
| mintLimitPerWallet | mintLimitPerWallet | Direct |
| stages[0].presale.price | mintPriceAllowlist | Convert to Wei |
| stages[0].public.price | mintPricePublic | Convert to Wei |
| mintStartAt | mintStartTime | ISO string |
| stages[0].presale.duration | allowlistStageDurationSeconds | Convert to seconds |
| stages[0].presale.allowlistAddresses | (separate mutation) | Use addToAllowlist |

### Testing Checklist

- [ ] Upload single image works
- [ ] Create ERC721 collection works
- [ ] Create ERC1155 collection works
- [ ] Allowlist addresses saved correctly
- [ ] Progress dialog shows correct states
- [ ] Error handling works
- [ ] Redirect after success works
- [ ] My Collections page loads
- [ ] Authentication required check works

---

**Total Implementation Steps**: 5 Phases
**Estimated Complexity**: Medium
**Dependencies**: Backend services must be running
