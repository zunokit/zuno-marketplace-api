package server

import (
	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConvertCollectionToProto converts a models.Collection to pb.Collection
func ConvertCollectionToProto(c *models.Collection) *pb.Collection {
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

	// Map optional fields
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

// MapCreateRequestToModel maps a CreateCollectionRequest to models.Collection
func MapCreateRequestToModel(req *pb.CreateCollectionRequest, userID uuid.UUID) *models.Collection {
	collection := &models.Collection{
		UserID:          userID,
		Name:            req.Name,
		Symbol:          req.Symbol,
		DeployerAddress: req.DeployerAddress,
		ImageURL:        req.ImageUrl,
	}

	// Map optional collection fields
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
		// Convert proto enum "TOKEN_STANDARD_ERC721" -> "ERC721"
		var ts models.TokenStandard
		switch req.TokenStandard {
		case pb.TokenStandard_TOKEN_STANDARD_ERC721:
			ts = models.TokenStandardERC721
		case pb.TokenStandard_TOKEN_STANDARD_ERC1155:
			ts = models.TokenStandardERC1155
		}
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

	// Map metadata
	collection.Metadata = MapMetadataFromRequest(req)

	return collection
}

// MapMetadataFromRequest maps metadata fields from CreateCollectionRequest
func MapMetadataFromRequest(req *pb.CreateCollectionRequest) *models.CollectionMetadata {
	metadata := &models.CollectionMetadata{}

	if req.MetadataUri != "" {
		metadata.MetadataURI = &req.MetadataUri
	}
	if req.IpfsHash != "" {
		metadata.IPFSHash = &req.IpfsHash
	}
	if req.DiscordUrl != "" {
		metadata.DiscordURL = &req.DiscordUrl
	}
	if req.TwitterUrl != "" {
		metadata.TwitterURL = &req.TwitterUrl
	}
	if req.InstagramUrl != "" {
		metadata.InstagramURL = &req.InstagramUrl
	}
	if req.MediumUrl != "" {
		metadata.MediumURL = &req.MediumUrl
	}
	if req.TelegramUrl != "" {
		metadata.TelegramURL = &req.TelegramUrl
	}

	return metadata
}

// MapUpdateRequestToUpdates maps UpdateCollectionRequest to update map
func MapUpdateRequestToUpdates(req *pb.UpdateCollectionRequest) map[string]interface{} {
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

	return updates
}

// MapMetadataUpdates maps metadata fields from UpdateCollectionRequest
func MapMetadataUpdates(req *pb.UpdateCollectionRequest) map[string]interface{} {
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

	return metadataUpdates
}
