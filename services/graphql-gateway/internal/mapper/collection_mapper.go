package mapper

import (
	"time"

	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/graph/model"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToGraphQLCollection converts a protobuf Collection to a GraphQL Collection
func ProtoToGraphQLCollection(proto *pb.Collection) *model.Collection {
	if proto == nil {
		return nil
	}

	collection := &model.Collection{
		ID:                 proto.Id,
		Slug:               ptrOrNil(proto.Slug),
		UserID:             proto.UserId,
		Name:               proto.Name,
		Symbol:             proto.Symbol,
		Description:        ptrOrNil(proto.Description),
		Category:           ptrOrNil(proto.Category),
		ContractAddress:    ptrOrNil(proto.ContractAddress),
		ChainID:            ptrOrNil(proto.ChainId),
		TokenStandard:      ProtoToGraphQLTokenStandard(proto.TokenStandard),
		DeployerAddress:    proto.DeployerAddress,
		DeployedBlock:      ptrIntOrNil(proto.DeployedBlock),
		Status:             ProtoToGraphQLCollectionStatus(proto.Status),
		DeployedAt:         timestampToTimePtrTime(proto.DeployedAt),
		IndexStatus:        ProtoToGraphQLIndexStatus(proto.IndexStatus),
		IsVerified:         proto.IsVerified,
		IsHidden:           proto.IsHidden,
		Source:             ptrOrNil(proto.Source),
		ImageURL:           ptrOrNil(proto.ImageUrl),
		BannerURL:          ptrOrNil(proto.BannerUrl),
		FeaturedImageURL:   ptrOrNil(proto.FeaturedImageUrl),
		WebsiteURL:         ptrOrNil(proto.WebsiteUrl),
		BaseURI:            ptrOrNil(proto.BaseUri),
		MaxSupply:          ptrIntOrNil(proto.MaxSupply),
		MintPriceAllowlist: ptrOrNil(proto.MintPriceAllowlist),
		MintPricePublic:    ptrOrNil(proto.MintPricePublic),
		MintStartTime:      timestampToTimePtrTime(proto.MintStartTime),
		AllowlistStageEnd:  timestampToTimePtrTime(proto.AllowlistStageEnd),
		MintLimitPerWallet: ptrInt32OrNil(proto.MintLimitPerWallet),
		RoyaltyFeeBps:      ptrInt32OrNil(proto.RoyaltyFeeBps),
		RoyaltyRecipient:   ptrOrNil(proto.RoyaltyRecipient),
		TotalSupply:        int(proto.TotalSupply),
		TotalMinted:        int(proto.TotalMinted),
		MetadataStandard:   ptrOrNil(proto.MetadataStandard),
		CreatedAt:          timestampToTimeObj(proto.CreatedAt),
		UpdatedAt:          timestampToTimeObj(proto.UpdatedAt),
		Metadata:           ProtoToGraphQLMetadata(proto.Metadata),
		Stats:              ProtoToGraphQLStats(proto.Stats),
	}

	return collection
}

// ProtoToGraphQLMetadata converts protobuf CollectionMetadata to GraphQL CollectionMetadata
func ProtoToGraphQLMetadata(proto *pb.CollectionMetadata) *model.CollectionMetadata {
	if proto == nil {
		return nil
	}

	return &model.CollectionMetadata{
		ID:              proto.Id,
		CollectionID:    proto.CollectionId,
		MetadataURI:     ptrOrNil(proto.MetadataUri),
		IpfsHash:        ptrOrNil(proto.IpfsHash),
		IpfsURL:         ptrOrNil(proto.IpfsUrl),
		DiscordURL:      ptrOrNil(proto.DiscordUrl),
		TwitterURL:      ptrOrNil(proto.TwitterUrl),
		InstagramURL:    ptrOrNil(proto.InstagramUrl),
		MediumURL:       ptrOrNil(proto.MediumUrl),
		TelegramURL:     ptrOrNil(proto.TelegramUrl),
		BackgroundColor: ptrOrNil(proto.BackgroundColor),
	}
}

// ProtoToGraphQLStats converts protobuf CollectionStats to GraphQL CollectionStats
func ProtoToGraphQLStats(proto *pb.CollectionStats) *model.CollectionStats {
	if proto == nil {
		return nil
	}

	return &model.CollectionStats{
		CollectionID:    proto.CollectionId,
		TotalItems:      int(proto.TotalItems),
		TotalOwners:     int(proto.TotalOwners),
		TotalSales:      int(proto.TotalSales),
		FloorPriceWei:   proto.FloorPriceWei,
		TotalVolumeWei:  proto.TotalVolumeWei,
		AveragePriceWei: proto.AveragePriceWei,
		Volume24hWei:    proto.Volume_24HWei,
		Sales24h:        int(proto.Sales_24H),
		LastSaleAt:      timestampToTimePtrTime(proto.LastSaleAt),
		LastMintAt:      timestampToTimePtrTime(proto.LastMintAt),
		UpdatedAt:       timestampToTimeObj(proto.UpdatedAt),
	}
}

// ProtoToGraphQLTokenStandard converts proto TokenStandard to GraphQL TokenStandard
func ProtoToGraphQLTokenStandard(proto pb.TokenStandard) model.TokenStandard {
	switch proto {
	case pb.TokenStandard_TOKEN_STANDARD_ERC721:
		return model.TokenStandardErc721
	case pb.TokenStandard_TOKEN_STANDARD_ERC1155:
		return model.TokenStandardErc1155
	default:
		return model.TokenStandardErc721
	}
}

// ProtoToGraphQLCollectionStatus converts proto CollectionStatus to GraphQL CollectionStatus
func ProtoToGraphQLCollectionStatus(proto pb.CollectionStatus) model.CollectionStatus {
	switch proto {
	case pb.CollectionStatus_COLLECTION_STATUS_PENDING:
		return model.CollectionStatusPending
	case pb.CollectionStatus_COLLECTION_STATUS_DEPLOYED:
		return model.CollectionStatusDeployed
	case pb.CollectionStatus_COLLECTION_STATUS_FAILED:
		return model.CollectionStatusFailed
	case pb.CollectionStatus_COLLECTION_STATUS_ARCHIVED:
		return model.CollectionStatusArchived
	default:
		return model.CollectionStatusPending
	}
}

// ProtoToGraphQLIndexStatus converts proto IndexStatus to GraphQL IndexStatus
func ProtoToGraphQLIndexStatus(proto pb.IndexStatus) *model.IndexStatus {
	if proto == pb.IndexStatus_INDEX_STATUS_UNSPECIFIED {
		return nil
	}

	var status model.IndexStatus
	switch proto {
	case pb.IndexStatus_INDEX_STATUS_NOT_STARTED:
		status = model.IndexStatusNotStarted
	case pb.IndexStatus_INDEX_STATUS_SYNCING:
		status = model.IndexStatusSyncing
	case pb.IndexStatus_INDEX_STATUS_SYNCED:
		status = model.IndexStatusSynced
	case pb.IndexStatus_INDEX_STATUS_FAILED:
		status = model.IndexStatusFailed
	default:
		status = model.IndexStatusNotStarted
	}
	return &status
}

// GraphQLToProtoTokenStandard converts GraphQL TokenStandard to proto TokenStandard
func GraphQLToProtoTokenStandard(gql model.TokenStandard) pb.TokenStandard {
	switch gql {
	case model.TokenStandardErc721:
		return pb.TokenStandard_TOKEN_STANDARD_ERC721
	case model.TokenStandardErc1155:
		return pb.TokenStandard_TOKEN_STANDARD_ERC1155
	default:
		return pb.TokenStandard_TOKEN_STANDARD_ERC721
	}
}

// GraphQLToProtoCollectionStatus converts GraphQL CollectionStatus to proto CollectionStatus
func GraphQLToProtoCollectionStatus(gql *model.CollectionStatus) pb.CollectionStatus {
	if gql == nil {
		return pb.CollectionStatus_COLLECTION_STATUS_UNSPECIFIED
	}

	switch *gql {
	case model.CollectionStatusPending:
		return pb.CollectionStatus_COLLECTION_STATUS_PENDING
	case model.CollectionStatusDeployed:
		return pb.CollectionStatus_COLLECTION_STATUS_DEPLOYED
	case model.CollectionStatusFailed:
		return pb.CollectionStatus_COLLECTION_STATUS_FAILED
	case model.CollectionStatusArchived:
		return pb.CollectionStatus_COLLECTION_STATUS_ARCHIVED
	default:
		return pb.CollectionStatus_COLLECTION_STATUS_PENDING
	}
}

// GraphQLCreateInputToProto converts GraphQL CreateCollectionInput to proto CreateCollectionRequest
func GraphQLCreateInputToProto(input model.CreateCollectionInput, userID string) *pb.CreateCollectionRequest {
	req := &pb.CreateCollectionRequest{
		UserId:             userID,
		Name:               input.Name,
		Symbol:             input.Symbol,
		Description:        stringPtrToValue(input.Description),
		Category:           stringPtrToValue(input.Category),
		TokenStandard:      GraphQLToProtoTokenStandard(input.TokenStandard),
		DeployerAddress:    input.DeployerAddress,
		ChainId:            stringPtrToValue(input.ChainID),
		ImageUrl:           stringPtrToValue(input.ImageURL),
		BannerUrl:          stringPtrToValue(input.BannerURL),
		FeaturedImageUrl:   stringPtrToValue(input.FeaturedImageURL),
		WebsiteUrl:         stringPtrToValue(input.WebsiteURL),
		RoyaltyFeeBps:      int32PtrToValue(input.RoyaltyFeeBps),
		RoyaltyRecipient:   stringPtrToValue(input.RoyaltyRecipient),
		BaseUri:            stringPtrToValue(input.BaseURI),
		MaxSupply:          int64PtrToValue(input.MaxSupply),
		MintPriceAllowlist: stringPtrToValue(input.MintPriceAllowlist),
		MintPricePublic:    stringPtrToValue(input.MintPricePublic),
		MintStartTime:      timePtrToTimestamp(input.MintStartTime),
		AllowlistStageEnd:  timePtrToTimestamp(input.AllowlistStageEnd),
		MintLimitPerWallet: int32PtrToValue(input.MintLimitPerWallet),
		MetadataUri:        stringPtrToValue(input.MetadataURI),
		IpfsHash:           stringPtrToValue(input.IpfsHash),
		DiscordUrl:         stringPtrToValue(input.DiscordURL),
		TwitterUrl:         stringPtrToValue(input.TwitterURL),
		InstagramUrl:       stringPtrToValue(input.InstagramURL),
		MediumUrl:          stringPtrToValue(input.MediumURL),
		TelegramUrl:        stringPtrToValue(input.TelegramURL),
	}

	return req
}

// GraphQLUpdateInputToProto converts GraphQL UpdateCollectionInput to proto UpdateCollectionRequest
func GraphQLUpdateInputToProto(id, userID string, input model.UpdateCollectionInput) *pb.UpdateCollectionRequest {
	req := &pb.UpdateCollectionRequest{
		Id:     id,
		UserId: userID,
	}

	// Optional fields
	if input.Description != nil {
		req.Description = input.Description
	}
	if input.Category != nil {
		req.Category = input.Category
	}
	if input.ImageURL != nil {
		req.ImageUrl = input.ImageURL
	}
	if input.BannerURL != nil {
		req.BannerUrl = input.BannerURL
	}
	if input.FeaturedImageURL != nil {
		req.FeaturedImageUrl = input.FeaturedImageURL
	}
	if input.WebsiteURL != nil {
		req.WebsiteUrl = input.WebsiteURL
	}
	if input.DiscordURL != nil {
		req.DiscordUrl = input.DiscordURL
	}
	if input.TwitterURL != nil {
		req.TwitterUrl = input.TwitterURL
	}
	if input.InstagramURL != nil {
		req.InstagramUrl = input.InstagramURL
	}
	if input.MediumURL != nil {
		req.MediumUrl = input.MediumURL
	}
	if input.TelegramURL != nil {
		req.TelegramUrl = input.TelegramURL
	}
	if input.BackgroundColor != nil {
		req.BackgroundColor = input.BackgroundColor
	}
	if input.BaseURI != nil {
		req.BaseUri = input.BaseURI
	}
	if input.MintPriceAllowlist != nil {
		req.MintPriceAllowlist = input.MintPriceAllowlist
	}
	if input.MintPricePublic != nil {
		req.MintPricePublic = input.MintPricePublic
	}
	if input.MintStartTime != nil {
		req.MintStartTime = timePtrToTimestamp(input.MintStartTime)
	}
	if input.AllowlistStageEnd != nil {
		req.AllowlistStageEnd = timePtrToTimestamp(input.AllowlistStageEnd)
	}
	if input.MintLimitPerWallet != nil {
		val := int32(*input.MintLimitPerWallet)
		req.MintLimitPerWallet = &val
	}
	if input.Status != nil {
		status := GraphQLToProtoCollectionStatus(input.Status)
		req.Status = &status
	}
	if input.ContractAddress != nil {
		req.ContractAddress = input.ContractAddress
	}
	if input.DeployedBlock != nil {
		deployedBlock := int64(*input.DeployedBlock)
		req.DeployedBlock = &deployedBlock
	}

	return req
}

// Helper functions

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrIntOrNil(i int64) *int {
	if i == 0 {
		return nil
	}
	val := int(i)
	return &val
}

func ptrInt32OrNil(i int32) *int {
	if i == 0 {
		return nil
	}
	val := int(i)
	return &val
}

func stringPtrToValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int32PtrToValue(i *int) int32 {
	if i == nil {
		return 0
	}
	return int32(*i)
}

func int64PtrToValue(i *int) int64 {
	if i == nil {
		return 0
	}
	return int64(*i)
}

func timestampToTimeObj(ts *timestamppb.Timestamp) time.Time {
	if ts == nil || !ts.IsValid() {
		return time.Time{}
	}
	return ts.AsTime()
}

func timestampToTimePtrTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func timePtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
