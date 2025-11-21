package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/models"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/repository"
	"github.com/quangdang46/NFT-Marketplace/services/collection-service/internal/service"
	"github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CollectionServer struct {
	pb.UnimplementedCollectionServiceServer
	service *service.CollectionService
}

func NewCollectionServer(service *service.CollectionService) *CollectionServer {
	return &CollectionServer{service: service}
}

// Helper to convert model to proto
func convertCollectionToProto(c *models.Collection) *pb.Collection {
	if c == nil {
		return nil
	}

	pbCollection := &pb.Collection{
		Id:              c.ID.String(),
		UserId:          c.UserID.String(),
		Name:            c.Name,
		Symbol:          c.Symbol,
		DeployerAddress: c.DeployerAddress,
		Status:          pb.CollectionStatus(pb.CollectionStatus_value["COLLECTION_STATUS_"+string(c.Status)]),
		IndexStatus:     pb.IndexStatus(pb.IndexStatus_value["INDEX_STATUS_"+string(c.IndexStatus)]),
		IsVerified:      c.IsVerified,
		IsHidden:        c.IsHidden,
		Source:          c.Source,
		ImageUrl:        c.ImageURL,
		TotalSupply:     int32(c.TotalSupply),
		TotalMinted:     int32(c.TotalMinted),
		CreatedAt:       timestamppb.New(c.CreatedAt),
		UpdatedAt:       timestamppb.New(c.UpdatedAt),
	}

	if c.Slug != nil {
		pbCollection.Slug = *c.Slug
	}
	if c.Description != nil {
		pbCollection.Description = *c.Description
	}
	if c.Category != nil {
		pbCollection.Category = *c.Category
	}
	if c.ContractAddress != nil {
		pbCollection.ContractAddress = *c.ContractAddress
	}
	if c.ChainID != nil {
		pbCollection.ChainId = *c.ChainID
	}
	if c.TokenStandard != nil {
		pbCollection.TokenStandard = pb.TokenStandard(pb.TokenStandard_value["TOKEN_STANDARD_"+string(*c.TokenStandard)])
	}
	if c.DeployedBlock != nil {
		pbCollection.DeployedBlock = *c.DeployedBlock
	}
	if c.DeployedAt != nil {
		pbCollection.DeployedAt = timestamppb.New(*c.DeployedAt)
	}
	if c.BannerURL != nil {
		pbCollection.BannerUrl = *c.BannerURL
	}
	if c.FeaturedImageURL != nil {
		pbCollection.FeaturedImageUrl = *c.FeaturedImageURL
	}
	if c.WebsiteURL != nil {
		pbCollection.WebsiteUrl = *c.WebsiteURL
	}
	if c.BaseURI != nil {
		pbCollection.BaseUri = *c.BaseURI
	}
	if c.MaxSupply != nil {
		pbCollection.MaxSupply = *c.MaxSupply
	}
	if c.MintPriceAllowlist != nil {
		pbCollection.MintPriceAllowlist = *c.MintPriceAllowlist
	}
	if c.MintPricePublic != nil {
		pbCollection.MintPricePublic = *c.MintPricePublic
	}
	if c.MintStartTime != nil {
		pbCollection.MintStartTime = timestamppb.New(*c.MintStartTime)
	}
	if c.AllowlistStageEnd != nil {
		pbCollection.AllowlistStageEnd = timestamppb.New(*c.AllowlistStageEnd)
	}
	if c.MintLimitPerWallet != nil {
		pbCollection.MintLimitPerWallet = int32(*c.MintLimitPerWallet)
	}
	if c.RoyaltyFeeBPS != nil {
		pbCollection.RoyaltyFeeBps = int32(*c.RoyaltyFeeBPS)
	}
	if c.RoyaltyRecipient != nil {
		pbCollection.RoyaltyRecipient = *c.RoyaltyRecipient
	}
	if c.MetadataStandard != nil {
		pbCollection.MetadataStandard = *c.MetadataStandard
	}

	// Metadata relations
	if c.Metadata != nil {
		pbCollection.Metadata = &pb.CollectionMetadata{
			Id:           c.Metadata.ID.String(),
			CollectionId: c.Metadata.CollectionID.String(),
		}
		if c.Metadata.MetadataURI != nil {
			pbCollection.Metadata.MetadataUri = *c.Metadata.MetadataURI
		}
		if c.Metadata.IPFSHash != nil {
			pbCollection.Metadata.IpfsHash = *c.Metadata.IPFSHash
		}
		if c.Metadata.IPFSURL != nil {
			pbCollection.Metadata.IpfsUrl = *c.Metadata.IPFSURL
		}
		if c.Metadata.DiscordURL != nil {
			pbCollection.Metadata.DiscordUrl = *c.Metadata.DiscordURL
		}
		if c.Metadata.TwitterURL != nil {
			pbCollection.Metadata.TwitterUrl = *c.Metadata.TwitterURL
		}
		if c.Metadata.InstagramURL != nil {
			pbCollection.Metadata.InstagramUrl = *c.Metadata.InstagramURL
		}
		if c.Metadata.MediumURL != nil {
			pbCollection.Metadata.MediumUrl = *c.Metadata.MediumURL
		}
		if c.Metadata.TelegramURL != nil {
			pbCollection.Metadata.TelegramUrl = *c.Metadata.TelegramURL
		}
		if c.Metadata.BackgroundColor != nil {
			pbCollection.Metadata.BackgroundColor = *c.Metadata.BackgroundColor
		}
	}

	// Stats relations
	if c.Stats != nil {
		pbCollection.Stats = &pb.CollectionStats{
			CollectionId:   c.Stats.CollectionID.String(),
			TotalItems:     int32(c.Stats.TotalItems),
			TotalOwners:    int32(c.Stats.TotalOwners),
			TotalSales:     int32(c.Stats.TotalSales),
			TotalVolumeWei: c.Stats.TotalVolumeWei,
			Volume_24HWei:  c.Stats.Volume24hWei,
			Sales_24H:      int32(c.Stats.Sales24h),
			UpdatedAt:      timestamppb.New(c.Stats.UpdatedAt),
		}
		if c.Stats.FloorPriceWei != nil {
			pbCollection.Stats.FloorPriceWei = *c.Stats.FloorPriceWei
		}
		if c.Stats.AveragePriceWei != nil {
			pbCollection.Stats.AveragePriceWei = *c.Stats.AveragePriceWei
		}
		if c.Stats.LastSaleAt != nil {
			pbCollection.Stats.LastSaleAt = timestamppb.New(*c.Stats.LastSaleAt)
		}
		if c.Stats.LastMintAt != nil {
			pbCollection.Stats.LastMintAt = timestamppb.New(*c.Stats.LastMintAt)
		}
	}

	return pbCollection
}

func (s *CollectionServer) CreateCollection(ctx context.Context, req *pb.CreateCollectionRequest) (*pb.CreateCollectionResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	// Map request to model
	collection := &models.Collection{
		UserID:          userID,
		Name:            req.Name,
		Symbol:          req.Symbol,
		DeployerAddress: req.DeployerAddress,
		ImageURL:        req.ImageUrl,
	}

	if req.Description != "" {
		collection.Description = &req.Description
	}
	if req.Category != "" {
		collection.Category = &req.Category
	}
	if req.ChainId != "" {
		collection.ChainID = &req.ChainId
	}
	if req.TokenStandard != pb.TokenStandard_TOKEN_STANDARD_UNSPECIFIED {
		ts := models.TokenStandard(req.TokenStandard.String())
		collection.TokenStandard = &ts
	}
	if req.BannerUrl != "" {
		collection.BannerURL = &req.BannerUrl
	}
	if req.FeaturedImageUrl != "" {
		collection.FeaturedImageURL = &req.FeaturedImageUrl
	}
	if req.WebsiteUrl != "" {
		collection.WebsiteURL = &req.WebsiteUrl
	}
	if req.BaseUri != "" {
		collection.BaseURI = &req.BaseUri
	}
	if req.MaxSupply > 0 {
		collection.MaxSupply = &req.MaxSupply
	}
	if req.MintPriceAllowlist != "" {
		collection.MintPriceAllowlist = &req.MintPriceAllowlist
	}
	if req.MintPricePublic != "" {
		collection.MintPricePublic = &req.MintPricePublic
	}
	if req.MintStartTime != nil {
		t := req.MintStartTime.AsTime()
		collection.MintStartTime = &t
	}
	if req.AllowlistStageEnd != nil {
		t := req.AllowlistStageEnd.AsTime()
		collection.AllowlistStageEnd = &t
	}
	if req.MintLimitPerWallet > 0 {
		limit := int(req.MintLimitPerWallet)
		collection.MintLimitPerWallet = &limit
	}
	if req.RoyaltyFeeBps >= 0 {
		fee := int(req.RoyaltyFeeBps)
		collection.RoyaltyFeeBPS = &fee
	}
	if req.RoyaltyRecipient != "" {
		collection.RoyaltyRecipient = &req.RoyaltyRecipient
	}

	// Metadata
	collection.Metadata = &models.CollectionMetadata{}
	if req.MetadataUri != "" {
		collection.Metadata.MetadataURI = &req.MetadataUri
	}
	if req.IpfsHash != "" {
		collection.Metadata.IPFSHash = &req.IpfsHash
	}
	if req.DiscordUrl != "" {
		collection.Metadata.DiscordURL = &req.DiscordUrl
	}
	if req.TwitterUrl != "" {
		collection.Metadata.TwitterURL = &req.TwitterUrl
	}
	if req.InstagramUrl != "" {
		collection.Metadata.InstagramURL = &req.InstagramUrl
	}
	if req.MediumUrl != "" {
		collection.Metadata.MediumURL = &req.MediumUrl
	}
	if req.TelegramUrl != "" {
		collection.Metadata.TelegramURL = &req.TelegramUrl
	}

	created, err := s.service.CreateCollection(ctx, collection)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCollectionResponse{
		Collection: convertCollectionToProto(created),
	}, nil
}

func (s *CollectionServer) GetCollection(ctx context.Context, req *pb.GetCollectionRequest) (*pb.GetCollectionResponse, error) {
	var collection *models.Collection
	var err error

	if req.Id != "" {
		id, parseErr := uuid.Parse(req.Id)
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid id")
		}
		collection, err = s.service.GetCollection(ctx, id)
	} else if req.ContractAddress != "" && req.ChainId != "" {
		collection, err = s.service.GetCollectionByContract(ctx, req.ContractAddress, req.ChainId)
	} else {
		return nil, status.Error(codes.InvalidArgument, "must provide id or contract_address+chain_id")
	}

	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetCollectionResponse{
		Collection: convertCollectionToProto(collection),
	}, nil
}

func (s *CollectionServer) UpdateCollection(ctx context.Context, req *pb.UpdateCollectionRequest) (*pb.UpdateCollectionResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	updates := make(map[string]interface{})

	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.ImageUrl != nil {
		updates["image_url"] = *req.ImageUrl
	}
	if req.BannerUrl != nil {
		updates["banner_url"] = *req.BannerUrl
	}
	if req.FeaturedImageUrl != nil {
		updates["featured_image_url"] = *req.FeaturedImageUrl
	}
	if req.WebsiteUrl != nil {
		updates["website_url"] = *req.WebsiteUrl
	}
	if req.BaseUri != nil {
		updates["base_uri"] = *req.BaseUri
	}
	if req.MintPriceAllowlist != nil {
		updates["mint_price_allowlist"] = *req.MintPriceAllowlist
	}
	if req.MintPricePublic != nil {
		updates["mint_price_public"] = *req.MintPricePublic
	}
	if req.MintStartTime != nil {
		updates["mint_start_time"] = req.MintStartTime.AsTime()
	}
	if req.AllowlistStageEnd != nil {
		updates["allowlist_stage_end"] = req.AllowlistStageEnd.AsTime()
	}
	if req.MintLimitPerWallet != nil {
		updates["mint_limit_per_wallet"] = int(*req.MintLimitPerWallet)
	}
	if req.Status != nil {
		updates["status"] = models.CollectionStatus(req.Status.String())
	}
	if req.ContractAddress != nil {
		updates["contract_address"] = *req.ContractAddress
	}
	if req.DeployedBlock != nil {
		updates["deployed_block"] = *req.DeployedBlock
	}

	// Handle metadata updates
	metadataUpdates := make(map[string]interface{})
	if req.DiscordUrl != nil {
		metadataUpdates["discord_url"] = *req.DiscordUrl
	}
	if req.TwitterUrl != nil {
		metadataUpdates["twitter_url"] = *req.TwitterUrl
	}
	if req.InstagramUrl != nil {
		metadataUpdates["instagram_url"] = *req.InstagramUrl
	}
	if req.MediumUrl != nil {
		metadataUpdates["medium_url"] = *req.MediumUrl
	}
	if req.TelegramUrl != nil {
		metadataUpdates["telegram_url"] = *req.TelegramUrl
	}
	if req.BackgroundColor != nil {
		metadataUpdates["background_color"] = *req.BackgroundColor
	}

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
		Collection: convertCollectionToProto(updated),
	}, nil
}

func (s *CollectionServer) ListCollectionsByUser(ctx context.Context, req *pb.ListCollectionsByUserRequest) (*pb.ListCollectionsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	limit := int(req.Limit)
	if limit < 1 {
		limit = 10
	}

	collections, total, err := s.service.ListCollectionsByUser(ctx, userID, page, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbCollections := make([]*pb.Collection, len(collections))
	for i, c := range collections {
		pbCollections[i] = convertCollectionToProto(c)
	}

	return &pb.ListCollectionsResponse{
		Collections: pbCollections,
		TotalCount:  total,
		Page:        int32(page),
		Limit:       int32(limit),
	}, nil
}

func (s *CollectionServer) ListCollections(ctx context.Context, req *pb.ListCollectionsRequest) (*pb.ListCollectionsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	limit := int(req.Limit)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

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

	collections, total, err := s.service.ListCollections(ctx, filters, page, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbCollections := make([]*pb.Collection, len(collections))
	for i, c := range collections {
		pbCollections[i] = convertCollectionToProto(c)
	}

	return &pb.ListCollectionsResponse{
		Collections: pbCollections,
		TotalCount:  total,
		Page:        int32(page),
		Limit:       int32(limit),
	}, nil
}

func (s *CollectionServer) AddToAllowlist(ctx context.Context, req *pb.AddToAllowlistRequest) (*pb.AddToAllowlistResponse, error) {
	collectionID, err := uuid.Parse(req.CollectionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid collection_id")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

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
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

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
