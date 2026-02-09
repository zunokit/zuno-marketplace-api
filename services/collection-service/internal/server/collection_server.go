package server

import (
	"context"

	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/repository"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/service"
	"github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CollectionServer struct {
	pb.UnimplementedCollectionServiceServer
	service            *service.CollectionService
	processedEventRepo repository.ProcessedEventRepository
	cacheRepo          repository.CollectionRepository // Cached wrapper for cache invalidation
	logger             *zap.Logger
}

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

func (s *CollectionServer) CreateCollection(ctx context.Context, req *pb.CreateCollectionRequest) (*pb.CreateCollectionResponse, error) {
	// Validate request
	if err := ValidateCreateCollectionRequest(req); err != nil {
		return nil, err
	}

	// Parse user ID
	userID, err := ValidateUUID(req.UserId, "user_id")
	if err != nil {
		return nil, err
	}

	// Map request to model
	collection := MapCreateRequestToModel(req, userID)

	// Create collection
	created, err := s.service.CreateCollection(ctx, collection)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCollectionResponse{
		Collection: ConvertCollectionToProto(created),
	}, nil
}

func (s *CollectionServer) GetCollection(ctx context.Context, req *pb.GetCollectionRequest) (*pb.GetCollectionResponse, error) {
	// Validate request
	if err := ValidateGetCollectionRequest(req); err != nil {
		return nil, err
	}

	var collection *models.Collection
	var err error

	// Get collection by ID or contract address
	if req.Id != "" {
		id, parseErr := ValidateUUID(req.Id, "id")
		if parseErr != nil {
			return nil, parseErr
		}
		collection, err = s.service.GetCollection(ctx, id)
	} else {
		collection, err = s.service.GetCollectionByContract(ctx, req.ContractAddress, req.ChainId)
	}

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetCollectionResponse{
		Collection: ConvertCollectionToProto(collection),
	}, nil
}

func (s *CollectionServer) UpdateCollection(ctx context.Context, req *pb.UpdateCollectionRequest) (*pb.UpdateCollectionResponse, error) {
	// Validate request
	if err := ValidateUpdateCollectionRequest(req); err != nil {
		return nil, err
	}

	// Parse IDs
	id, err := ValidateUUID(req.Id, "id")
	if err != nil {
		return nil, err
	}
	userID, err := ValidateUUID(req.UserId, "user_id")
	if err != nil {
		return nil, err
	}

	// Map updates
	updates := MapUpdateRequestToUpdates(req)
	metadataUpdates := MapMetadataUpdates(req)

	// Update collection
	updated, err := s.service.UpdateCollection(ctx, id, userID, updates)
	if err != nil {
		if err == service.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Update metadata if there are metadata changes
	if len(metadataUpdates) > 0 {
		if err := s.service.UpdateMetadata(ctx, id, userID, metadataUpdates); err != nil {
			if err == service.ErrUnauthorized {
				return nil, status.Error(codes.PermissionDenied, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.UpdateCollectionResponse{
		Collection: ConvertCollectionToProto(updated),
	}, nil
}

func (s *CollectionServer) ListCollectionsByUser(ctx context.Context, req *pb.ListCollectionsByUserRequest) (*pb.ListCollectionsResponse, error) {
	// Validate request
	if err := ValidateListCollectionsByUserRequest(req); err != nil {
		return nil, err
	}

	// Parse user ID
	userID, err := ValidateUUID(req.UserId, "user_id")
	if err != nil {
		return nil, err
	}

	// Validate pagination
	params := ValidatePaginationParams(req.Page, req.Limit, 0)

	// Get collections
	collections, total, err := s.service.ListCollectionsByUser(ctx, userID, params.Page, params.Limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to proto
	pbCollections := make([]*pb.Collection, len(collections))
	for i, c := range collections {
		pbCollections[i] = ConvertCollectionToProto(c)
	}

	return &pb.ListCollectionsResponse{
		Collections: pbCollections,
		TotalCount:  total,
		Page:        int32(params.Page),
		Limit:       int32(params.Limit),
	}, nil
}

func (s *CollectionServer) ListCollections(ctx context.Context, req *pb.ListCollectionsRequest) (*pb.ListCollectionsResponse, error) {
	// Validate pagination with max limit of 100
	params := ValidatePaginationParams(req.Page, req.Limit, 100)

	// Map filters from proto request
	filters := &repository.ListFilters{
		SortBy:      req.SortBy,
		SortOrder:   req.SortOrder,
		Category:    req.Category,
		ChainID:     req.ChainId,
		SearchQuery: req.SearchQuery,
	}

	if req.IsVerified != nil {
		isVerified := *req.IsVerified
		filters.IsVerified = &isVerified
	}

	// Get collections
	collections, total, err := s.service.ListCollections(ctx, filters, params.Page, params.Limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to proto
	pbCollections := make([]*pb.Collection, len(collections))
	for i, c := range collections {
		pbCollections[i] = ConvertCollectionToProto(c)
	}

	return &pb.ListCollectionsResponse{
		Collections: pbCollections,
		TotalCount:  total,
		Page:        int32(params.Page),
		Limit:       int32(params.Limit),
	}, nil
}

func (s *CollectionServer) AddToAllowlist(ctx context.Context, req *pb.AddToAllowlistRequest) (*pb.AddToAllowlistResponse, error) {
	// Validate request
	if err := ValidateAddToAllowlistRequest(req); err != nil {
		return nil, err
	}

	// Parse IDs
	collectionID, err := ValidateUUID(req.CollectionId, "collection_id")
	if err != nil {
		return nil, err
	}
	userID, err := ValidateUUID(req.UserId, "user_id")
	if err != nil {
		return nil, err
	}

	// Add to allowlist
	count, err := s.service.AddToAllowlist(ctx, collectionID, userID, req.WalletAddresses, int(req.MaxMintAmount))
	if err != nil {
		if err == service.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddToAllowlistResponse{
		AddedCount: int32(count),
	}, nil
}

func (s *CollectionServer) DeleteCollection(ctx context.Context, req *pb.DeleteCollectionRequest) (*pb.DeleteCollectionResponse, error) {
	// Validate request
	if err := ValidateDeleteCollectionRequest(req); err != nil {
		return nil, err
	}

	// Parse IDs
	id, err := ValidateUUID(req.Id, "id")
	if err != nil {
		return nil, err
	}
	userID, err := ValidateUUID(req.UserId, "user_id")
	if err != nil {
		return nil, err
	}

	// Delete collection
	err = s.service.DeleteCollection(ctx, id, userID)
	if err != nil {
		if err == service.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteCollectionResponse{
		Success: true,
	}, nil
}
