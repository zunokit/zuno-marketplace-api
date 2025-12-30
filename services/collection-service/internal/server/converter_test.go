package server

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/services/collection-service/internal/models"
	"github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConvertCollectionToProto(t *testing.T) {
	now := time.Now()
	collectionID := uuid.New()
	userID := uuid.New()
	metadataID := uuid.New()

	slug := "test-collection"
	description := "Test description"
	category := "Art"
	contractAddress := "0x1234567890123456789012345678901234567890"
	chainID := "eip155:1"
	tokenStandard := models.TokenStandardERC721
	deployedBlock := int64(12345)
	bannerURL := "https://example.com/banner.png"
	featuredImageURL := "https://example.com/featured.png"
	websiteURL := "https://example.com"
	baseURI := "https://api.example.com/metadata/"
	maxSupply := int64(10000)
	mintPriceAllowlist := "0.05"
	mintPricePublic := "0.1"
	mintStartTime := now.Add(24 * time.Hour)
	allowlistStageEnd := now.Add(48 * time.Hour)
	mintLimitPerWallet := 5
	royaltyFeeBPS := 500
	royaltyRecipient := "0xrecipient"
	metadataStandard := "ERC721Metadata"

	floorPrice := "1.5"
	avgPrice := "2.0"
	lastSaleAt := now.Add(-1 * time.Hour)
	lastMintAt := now.Add(-30 * time.Minute)

	metadataURI := "https://metadata.example.com"
	ipfsHash := "QmHash123"
	ipfsURL := "ipfs://QmHash123"
	discordURL := "https://discord.gg/test"
	twitterURL := "https://twitter.com/test"
	instagramURL := "https://instagram.com/test"
	mediumURL := "https://medium.com/@test"
	telegramURL := "https://t.me/test"
	backgroundColor := "#FFFFFF"

	tests := []struct {
		name       string
		collection *models.Collection
		want       *pb.Collection
	}{
		{
			name:       "nil collection returns nil",
			collection: nil,
			want:       nil,
		},
		{
			name: "collection with all fields",
			collection: &models.Collection{
				ID:                 collectionID,
				UserID:             userID,
				Name:               "Test Collection",
				Symbol:             "TEST",
				Slug:               &slug,
				Description:        &description,
				Category:           &category,
				DeployerAddress:    "0xdeployer",
				ContractAddress:    &contractAddress,
				ChainID:            &chainID,
				TokenStandard:      &tokenStandard,
				DeployedBlock:      &deployedBlock,
				DeployedAt:         &now,
				Status:             models.CollectionStatusPending,
				IndexStatus:        models.IndexStatusNotStarted,
				IsVerified:         false,
				IsHidden:           false,
				Source:             "manual",
				ImageURL:           "https://example.com/image.png",
				BannerURL:          &bannerURL,
				FeaturedImageURL:   &featuredImageURL,
				WebsiteURL:         &websiteURL,
				BaseURI:            &baseURI,
				MaxSupply:          &maxSupply,
				TotalSupply:        1000,
				TotalMinted:        500,
				MintPriceAllowlist: &mintPriceAllowlist,
				MintPricePublic:    &mintPricePublic,
				MintStartTime:      &mintStartTime,
				AllowlistStageEnd:  &allowlistStageEnd,
				MintLimitPerWallet: &mintLimitPerWallet,
				RoyaltyFeeBPS:      &royaltyFeeBPS,
				RoyaltyRecipient:   &royaltyRecipient,
				MetadataStandard:   &metadataStandard,
				Metadata: &models.CollectionMetadata{
					ID:              metadataID,
					CollectionID:    collectionID,
					MetadataURI:     &metadataURI,
					IPFSHash:        &ipfsHash,
					IPFSURL:         &ipfsURL,
					DiscordURL:      &discordURL,
					TwitterURL:      &twitterURL,
					InstagramURL:    &instagramURL,
					MediumURL:       &mediumURL,
					TelegramURL:     &telegramURL,
					BackgroundColor: &backgroundColor,
					CreatedAt:       now,
					UpdatedAt:       now,
				},
				Stats: &models.CollectionStats{
					CollectionID:    collectionID,
					TotalItems:      1000,
					TotalOwners:     250,
					FloorPriceWei:   &floorPrice,
					AveragePriceWei: &avgPrice,
					TotalSales:      100,
					TotalVolumeWei:  "150.0",
					Volume24hWei:    "5.0",
					Sales24h:        10,
					LastSaleAt:      &lastSaleAt,
					LastMintAt:      &lastMintAt,
					UpdatedAt:       now,
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			want: &pb.Collection{
				Id:                 collectionID.String(),
				UserId:             userID.String(),
				Name:               "Test Collection",
				Symbol:             "TEST",
				Slug:               slug,
				Description:        description,
				Category:           category,
				DeployerAddress:    "0xdeployer",
				ContractAddress:    contractAddress,
				ChainId:            chainID,
				TokenStandard:      pb.TokenStandard_TOKEN_STANDARD_ERC721,
				DeployedBlock:      deployedBlock,
				DeployedAt:         timestamppb.New(now),
				Status:             pb.CollectionStatus_COLLECTION_STATUS_PENDING,
				IndexStatus:        pb.IndexStatus_INDEX_STATUS_NOT_STARTED,
				IsVerified:         false,
				IsHidden:           false,
				Source:             "manual",
				ImageUrl:           "https://example.com/image.png",
				BannerUrl:          bannerURL,
				FeaturedImageUrl:   featuredImageURL,
				WebsiteUrl:         websiteURL,
				BaseUri:            baseURI,
				MaxSupply:          maxSupply,
				TotalSupply:        1000,
				TotalMinted:        500,
				MintPriceAllowlist: mintPriceAllowlist,
				MintPricePublic:    mintPricePublic,
				MintStartTime:      timestamppb.New(mintStartTime),
				AllowlistStageEnd:  timestamppb.New(allowlistStageEnd),
				MintLimitPerWallet: int32(mintLimitPerWallet),
				RoyaltyFeeBps:      int32(royaltyFeeBPS),
				RoyaltyRecipient:   royaltyRecipient,
				MetadataStandard:   metadataStandard,
				Metadata: &pb.CollectionMetadata{
					Id:              metadataID.String(),
					CollectionId:    collectionID.String(),
					MetadataUri:     metadataURI,
					IpfsHash:        ipfsHash,
					IpfsUrl:         ipfsURL,
					DiscordUrl:      discordURL,
					TwitterUrl:      twitterURL,
					InstagramUrl:    instagramURL,
					MediumUrl:       mediumURL,
					TelegramUrl:     telegramURL,
					BackgroundColor: backgroundColor,
				},
				Stats: &pb.CollectionStats{
					CollectionId:    collectionID.String(),
					TotalItems:      1000,
					TotalOwners:     250,
					FloorPriceWei:   floorPrice,
					AveragePriceWei: avgPrice,
					TotalSales:      100,
					TotalVolumeWei:  "150.0",
					Volume_24HWei:   "5.0",
					Sales_24H:       10,
					LastSaleAt:      timestamppb.New(lastSaleAt),
					LastMintAt:      timestamppb.New(lastMintAt),
					UpdatedAt:       timestamppb.New(now),
				},
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
		},
		{
			name: "collection with minimal fields",
			collection: &models.Collection{
				ID:              collectionID,
				UserID:          userID,
				Name:            "Minimal Collection",
				Symbol:          "MIN",
				DeployerAddress: "0xdeployer",
				Status:          models.CollectionStatusPending,
				IndexStatus:     models.IndexStatusNotStarted,
				IsVerified:      false,
				IsHidden:        false,
				Source:          "manual",
				ImageURL:        "https://example.com/image.png",
				TotalSupply:     0,
				TotalMinted:     0,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			want: &pb.Collection{
				Id:              collectionID.String(),
				UserId:          userID.String(),
				Name:            "Minimal Collection",
				Symbol:          "MIN",
				DeployerAddress: "0xdeployer",
				Status:          pb.CollectionStatus_COLLECTION_STATUS_PENDING,
				IndexStatus:     pb.IndexStatus_INDEX_STATUS_NOT_STARTED,
				IsVerified:      false,
				IsHidden:        false,
				Source:          "manual",
				ImageUrl:        "https://example.com/image.png",
				TotalSupply:     0,
				TotalMinted:     0,
				CreatedAt:       timestamppb.New(now),
				UpdatedAt:       timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertCollectionToProto(tt.collection)

			if got == nil && tt.want == nil {
				return
			}

			if got == nil || tt.want == nil {
				t.Errorf("ConvertCollectionToProto() got nil mismatch, got = %v, want = %v", got == nil, tt.want == nil)
				return
			}

			// Compare basic fields
			if got.Id != tt.want.Id {
				t.Errorf("ConvertCollectionToProto().Id = %v, want %v", got.Id, tt.want.Id)
			}
			if got.Name != tt.want.Name {
				t.Errorf("ConvertCollectionToProto().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Symbol != tt.want.Symbol {
				t.Errorf("ConvertCollectionToProto().Symbol = %v, want %v", got.Symbol, tt.want.Symbol)
			}
		})
	}
}

func TestMapCreateRequestToModel(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name         string
		req          *pb.CreateCollectionRequest
		userID       uuid.UUID
		validateFunc func(*testing.T, *models.Collection)
	}{
		{
			name: "request with all fields",
			req: &pb.CreateCollectionRequest{
				UserId:             userID.String(),
				Name:               "Test Collection",
				Symbol:             "TEST",
				Description:        "Test description",
				Category:           "Art",
				DeployerAddress:    "0xdeployer",
				ChainId:            "eip155:1",
				TokenStandard:      pb.TokenStandard_TOKEN_STANDARD_ERC721,
				ImageUrl:           "https://example.com/image.png",
				BannerUrl:          "https://example.com/banner.png",
				FeaturedImageUrl:   "https://example.com/featured.png",
				WebsiteUrl:         "https://example.com",
				BaseUri:            "https://api.example.com/",
				MaxSupply:          10000,
				MintPriceAllowlist: "0.05",
				MintPricePublic:    "0.1",
				MintStartTime:      timestamppb.New(now),
				AllowlistStageEnd:  timestamppb.New(now.Add(24 * time.Hour)),
				MintLimitPerWallet: 5,
				RoyaltyFeeBps:      500,
				RoyaltyRecipient:   "0xrecipient",
				MetadataUri:        "https://metadata.example.com",
				IpfsHash:           "QmHash123",
				DiscordUrl:         "https://discord.gg/test",
				TwitterUrl:         "https://twitter.com/test",
				InstagramUrl:       "https://instagram.com/test",
				MediumUrl:          "https://medium.com/@test",
				TelegramUrl:        "https://t.me/test",
			},
			userID: userID,
			validateFunc: func(t *testing.T, c *models.Collection) {
				if c.UserID != userID {
					t.Errorf("UserID = %v, want %v", c.UserID, userID)
				}
				if c.Name != "Test Collection" {
					t.Errorf("Name = %v, want Test Collection", c.Name)
				}
				if c.Symbol != "TEST" {
					t.Errorf("Symbol = %v, want TEST", c.Symbol)
				}
				if c.Description == nil || *c.Description != "Test description" {
					t.Errorf("Description = %v, want Test description", c.Description)
				}
				if c.Category == nil || *c.Category != "Art" {
					t.Errorf("Category = %v, want Art", c.Category)
				}
				if c.ChainID == nil || *c.ChainID != "eip155:1" {
					t.Errorf("ChainID = %v, want eip155:1", c.ChainID)
				}
				if c.Metadata == nil {
					t.Error("Metadata should not be nil")
				}
			},
		},
		{
			name: "request with minimal fields",
			req: &pb.CreateCollectionRequest{
				UserId:          userID.String(),
				Name:            "Minimal Collection",
				Symbol:          "MIN",
				DeployerAddress: "0xdeployer",
				ImageUrl:        "https://example.com/image.png",
			},
			userID: userID,
			validateFunc: func(t *testing.T, c *models.Collection) {
				if c.UserID != userID {
					t.Errorf("UserID = %v, want %v", c.UserID, userID)
				}
				if c.Name != "Minimal Collection" {
					t.Errorf("Name = %v, want Minimal Collection", c.Name)
				}
				if c.Description != nil {
					t.Errorf("Description should be nil, got %v", c.Description)
				}
				if c.Category != nil {
					t.Errorf("Category should be nil, got %v", c.Category)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapCreateRequestToModel(tt.req, tt.userID)
			tt.validateFunc(t, got)
		})
	}
}

func TestMapMetadataFromRequest(t *testing.T) {
	tests := []struct {
		name         string
		req          *pb.CreateCollectionRequest
		validateFunc func(*testing.T, *models.CollectionMetadata)
	}{
		{
			name: "request with all metadata fields",
			req: &pb.CreateCollectionRequest{
				MetadataUri:  "https://metadata.example.com",
				IpfsHash:     "QmHash123",
				DiscordUrl:   "https://discord.gg/test",
				TwitterUrl:   "https://twitter.com/test",
				InstagramUrl: "https://instagram.com/test",
				MediumUrl:    "https://medium.com/@test",
				TelegramUrl:  "https://t.me/test",
			},
			validateFunc: func(t *testing.T, m *models.CollectionMetadata) {
				if m.MetadataURI == nil || *m.MetadataURI != "https://metadata.example.com" {
					t.Errorf("MetadataURI = %v, want https://metadata.example.com", m.MetadataURI)
				}
				if m.IPFSHash == nil || *m.IPFSHash != "QmHash123" {
					t.Errorf("IPFSHash = %v, want QmHash123", m.IPFSHash)
				}
				if m.DiscordURL == nil || *m.DiscordURL != "https://discord.gg/test" {
					t.Errorf("DiscordURL = %v, want https://discord.gg/test", m.DiscordURL)
				}
			},
		},
		{
			name: "request with no metadata fields",
			req:  &pb.CreateCollectionRequest{},
			validateFunc: func(t *testing.T, m *models.CollectionMetadata) {
				if m.MetadataURI != nil {
					t.Errorf("MetadataURI should be nil, got %v", m.MetadataURI)
				}
				if m.IPFSHash != nil {
					t.Errorf("IPFSHash should be nil, got %v", m.IPFSHash)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapMetadataFromRequest(tt.req)
			tt.validateFunc(t, got)
		})
	}
}

func TestMapUpdateRequestToUpdates(t *testing.T) {
	desc := "Updated description"
	category := "Gaming"
	imageURL := "https://example.com/new-image.png"

	tests := []struct {
		name string
		req  *pb.UpdateCollectionRequest
		want map[string]interface{}
	}{
		{
			name: "request with multiple updates",
			req: &pb.UpdateCollectionRequest{
				Description: &desc,
				Category:    &category,
				ImageUrl:    &imageURL,
			},
			want: map[string]interface{}{
				"description": desc,
				"category":    category,
				"image_url":   imageURL,
			},
		},
		{
			name: "request with no updates",
			req:  &pb.UpdateCollectionRequest{},
			want: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapUpdateRequestToUpdates(tt.req)

			if len(got) != len(tt.want) {
				t.Errorf("MapUpdateRequestToUpdates() length = %v, want %v", len(got), len(tt.want))
			}

			for key, wantValue := range tt.want {
				gotValue, exists := got[key]
				if !exists {
					t.Errorf("MapUpdateRequestToUpdates() missing key %v", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("MapUpdateRequestToUpdates()[%v] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestMapMetadataUpdates(t *testing.T) {
	discordURL := "https://discord.gg/updated"
	twitterURL := "https://twitter.com/updated"
	backgroundColor := "#000000"

	tests := []struct {
		name string
		req  *pb.UpdateCollectionRequest
		want map[string]interface{}
	}{
		{
			name: "request with multiple metadata updates",
			req: &pb.UpdateCollectionRequest{
				DiscordUrl:      &discordURL,
				TwitterUrl:      &twitterURL,
				BackgroundColor: &backgroundColor,
			},
			want: map[string]interface{}{
				"discord_url":      discordURL,
				"twitter_url":      twitterURL,
				"background_color": backgroundColor,
			},
		},
		{
			name: "request with no metadata updates",
			req:  &pb.UpdateCollectionRequest{},
			want: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapMetadataUpdates(tt.req)

			if len(got) != len(tt.want) {
				t.Errorf("MapMetadataUpdates() length = %v, want %v", len(got), len(tt.want))
			}

			for key, wantValue := range tt.want {
				gotValue, exists := got[key]
				if !exists {
					t.Errorf("MapMetadataUpdates() missing key %v", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("MapMetadataUpdates()[%v] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}
