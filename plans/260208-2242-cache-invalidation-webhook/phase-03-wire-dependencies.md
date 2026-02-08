# Phase 03: Wire Dependencies

## Overview
Update server struct and main.go to wire the cached repository.

## Files to Modify

### 1. collection_server.go

Add `cacheRepo` field to CollectionServer:

```go
type CollectionServer struct {
	pb.UnimplementedCollectionServiceServer
	service            *service.CollectionService
	processedEventRepo repository.ProcessedEventRepository
	cacheRepo          repository.CollectionRepository // Cached wrapper
	logger             *zap.Logger
}
```

Update constructor:

```go
func NewCollectionServer(
	service *service.CollectionService,
	processedEventRepo repository.ProcessedEventRepository,
	cacheRepo repository.CollectionRepository,
	logger *zap.Logger,
) *CollectionServer {
	return &CollectionServer{
		service:            service,
		processedEventRepo: processedEventRepo,
		cacheRepo:          cacheRepo,
		logger:             logger,
	}
}
```

### 2. main.go

Update repository initialization:

```go
// Initialize repositories
baseCollectionRepo := repository.NewCollectionRepository(db)
collectionRepo := repository.NewCachedCollectionRepository(baseCollectionRepo)
allowlistRepo := repository.NewAllowlistRepository(db)
metadataRepo := repository.NewMetadataRepository(db)
processedEventRepo := repository.NewProcessedEventRepository(db)
```

Update server initialization:

```go
collectionServer := server.NewCollectionServer(collectionSvc, processedEventRepo, collectionRepo, zapLogger)
```

## Success Criteria
- [ ] CollectionServer has cacheRepo field
- [ ] Constructor accepts cacheRepo parameter
- [ ] main.go wraps repository with cache
- [ ] Application compiles and runs
