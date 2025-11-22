package mapper

import (
	"testing"
	"time"

	"github.com/quangdang46/NFT-Marketplace/services/graphql-gateway/graph/model"
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProtoToGraphQLTokenStandard(t *testing.T) {
	tests := []struct {
		name     string
		input    pb.TokenStandard
		expected model.TokenStandard
	}{
		{
			name:     "ERC721",
			input:    pb.TokenStandard_TOKEN_STANDARD_ERC721,
			expected: model.TokenStandardErc721,
		},
		{
			name:     "ERC1155",
			input:    pb.TokenStandard_TOKEN_STANDARD_ERC1155,
			expected: model.TokenStandardErc1155,
		},
		{
			name:     "Unspecified defaults to ERC721",
			input:    pb.TokenStandard_TOKEN_STANDARD_UNSPECIFIED,
			expected: model.TokenStandardErc721,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoToGraphQLTokenStandard(tt.input)
			if result != tt.expected {
				t.Errorf("ProtoToGraphQLTokenStandard() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProtoToGraphQLCollectionStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    pb.CollectionStatus
		expected model.CollectionStatus
	}{
		{
			name:     "Pending",
			input:    pb.CollectionStatus_COLLECTION_STATUS_PENDING,
			expected: model.CollectionStatusPending,
		},
		{
			name:     "Deployed",
			input:    pb.CollectionStatus_COLLECTION_STATUS_DEPLOYED,
			expected: model.CollectionStatusDeployed,
		},
		{
			name:     "Failed",
			input:    pb.CollectionStatus_COLLECTION_STATUS_FAILED,
			expected: model.CollectionStatusFailed,
		},
		{
			name:     "Archived",
			input:    pb.CollectionStatus_COLLECTION_STATUS_ARCHIVED,
			expected: model.CollectionStatusArchived,
		},
		{
			name:     "Unspecified defaults to Pending",
			input:    pb.CollectionStatus_COLLECTION_STATUS_UNSPECIFIED,
			expected: model.CollectionStatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoToGraphQLCollectionStatus(tt.input)
			if result != tt.expected {
				t.Errorf("ProtoToGraphQLCollectionStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGraphQLToProtoTokenStandard(t *testing.T) {
	tests := []struct {
		name     string
		input    model.TokenStandard
		expected pb.TokenStandard
	}{
		{
			name:     "ERC721",
			input:    model.TokenStandardErc721,
			expected: pb.TokenStandard_TOKEN_STANDARD_ERC721,
		},
		{
			name:     "ERC1155",
			input:    model.TokenStandardErc1155,
			expected: pb.TokenStandard_TOKEN_STANDARD_ERC1155,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GraphQLToProtoTokenStandard(tt.input)
			if result != tt.expected {
				t.Errorf("GraphQLToProtoTokenStandard() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGraphQLToProtoCollectionStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    *model.CollectionStatus
		expected pb.CollectionStatus
	}{
		{
			name:     "Pending",
			input:    ptrCollectionStatus(model.CollectionStatusPending),
			expected: pb.CollectionStatus_COLLECTION_STATUS_PENDING,
		},
		{
			name:     "Deployed",
			input:    ptrCollectionStatus(model.CollectionStatusDeployed),
			expected: pb.CollectionStatus_COLLECTION_STATUS_DEPLOYED,
		},
		{
			name:     "Failed",
			input:    ptrCollectionStatus(model.CollectionStatusFailed),
			expected: pb.CollectionStatus_COLLECTION_STATUS_FAILED,
		},
		{
			name:     "Archived",
			input:    ptrCollectionStatus(model.CollectionStatusArchived),
			expected: pb.CollectionStatus_COLLECTION_STATUS_ARCHIVED,
		},
		{
			name:     "Nil returns unspecified",
			input:    nil,
			expected: pb.CollectionStatus_COLLECTION_STATUS_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GraphQLToProtoCollectionStatus(tt.input)
			if result != tt.expected {
				t.Errorf("GraphQLToProtoCollectionStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProtoToGraphQLCollection(t *testing.T) {
	now := time.Now()
	nowProto := timestamppb.New(now)

	tests := []struct {
		name     string
		input    *pb.Collection
		expected *model.Collection
	}{
		{
			name:     "Nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name: "Full collection conversion",
			input: &pb.Collection{
				Id:              "test-id",
				UserId:          "user-123",
				Name:            "Test Collection",
				Symbol:          "TEST",
				Description:     "A test collection",
				Category:        "art",
				ContractAddress: "0x1234567890123456789012345678901234567890",
				ChainId:         "eip155:1",
				TokenStandard:   pb.TokenStandard_TOKEN_STANDARD_ERC721,
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				DeployedBlock:   12345,
				Status:          pb.CollectionStatus_COLLECTION_STATUS_DEPLOYED,
				DeployedAt:      nowProto,
				IsVerified:      true,
				IsHidden:        false,
				ImageUrl:        "https://example.com/image.png",
				TotalSupply:     100,
				TotalMinted:     50,
				CreatedAt:       nowProto,
				UpdatedAt:       nowProto,
			},
			expected: &model.Collection{
				ID:              "test-id",
				UserID:          "user-123",
				Name:            "Test Collection",
				Symbol:          "TEST",
				Description:     ptrString("A test collection"),
				Category:        ptrString("art"),
				ContractAddress: ptrString("0x1234567890123456789012345678901234567890"),
				ChainID:         ptrString("eip155:1"),
				TokenStandard:   model.TokenStandardErc721,
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				DeployedBlock:   ptrInt(12345),
				Status:          model.CollectionStatusDeployed,
				DeployedAt:      &now,
				IsVerified:      true,
				IsHidden:        false,
				ImageURL:        ptrString("https://example.com/image.png"),
				TotalSupply:     100,
				TotalMinted:     50,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoToGraphQLCollection(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("ProtoToGraphQLCollection() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("ProtoToGraphQLCollection() = nil, want non-nil")
			}

			// Check basic fields
			if result.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", result.ID, tt.expected.ID)
			}
			if result.UserID != tt.expected.UserID {
				t.Errorf("UserID = %v, want %v", result.UserID, tt.expected.UserID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
			}
			if result.Symbol != tt.expected.Symbol {
				t.Errorf("Symbol = %v, want %v", result.Symbol, tt.expected.Symbol)
			}
			if result.TokenStandard != tt.expected.TokenStandard {
				t.Errorf("TokenStandard = %v, want %v", result.TokenStandard, tt.expected.TokenStandard)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Status = %v, want %v", result.Status, tt.expected.Status)
			}
			if result.IsVerified != tt.expected.IsVerified {
				t.Errorf("IsVerified = %v, want %v", result.IsVerified, tt.expected.IsVerified)
			}
			if result.TotalSupply != tt.expected.TotalSupply {
				t.Errorf("TotalSupply = %v, want %v", result.TotalSupply, tt.expected.TotalSupply)
			}
		})
	}
}

func TestProtoToGraphQLMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.CollectionMetadata
		expected *model.CollectionMetadata
	}{
		{
			name:     "Nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name: "Full metadata conversion",
			input: &pb.CollectionMetadata{
				Id:              "meta-id",
				CollectionId:    "collection-id",
				MetadataUri:     "ipfs://QmTest",
				IpfsHash:        "QmTest",
				IpfsUrl:         "https://ipfs.io/ipfs/QmTest",
				DiscordUrl:      "https://discord.gg/test",
				TwitterUrl:      "https://twitter.com/test",
				InstagramUrl:    "https://instagram.com/test",
				BackgroundColor: "#000000",
			},
			expected: &model.CollectionMetadata{
				ID:              "meta-id",
				CollectionID:    "collection-id",
				MetadataURI:     ptrString("ipfs://QmTest"),
				IpfsHash:        ptrString("QmTest"),
				IpfsURL:         ptrString("https://ipfs.io/ipfs/QmTest"),
				DiscordURL:      ptrString("https://discord.gg/test"),
				TwitterURL:      ptrString("https://twitter.com/test"),
				InstagramURL:    ptrString("https://instagram.com/test"),
				BackgroundColor: ptrString("#000000"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoToGraphQLMetadata(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("ProtoToGraphQLMetadata() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("ProtoToGraphQLMetadata() = nil, want non-nil")
			}

			if result.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", result.ID, tt.expected.ID)
			}
			if result.CollectionID != tt.expected.CollectionID {
				t.Errorf("CollectionID = %v, want %v", result.CollectionID, tt.expected.CollectionID)
			}
		})
	}
}

func TestProtoToGraphQLStats(t *testing.T) {
	now := time.Now()
	nowProto := timestamppb.New(now)

	tests := []struct {
		name     string
		input    *pb.CollectionStats
		expected *model.CollectionStats
	}{
		{
			name:     "Nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name: "Full stats conversion",
			input: &pb.CollectionStats{
				CollectionId:    "collection-id",
				TotalItems:      100,
				TotalOwners:     50,
				TotalSales:      25,
				FloorPriceWei:   "1000000000000000000",
				TotalVolumeWei:  "25000000000000000000",
				AveragePriceWei: "1000000000000000000",
				Volume_24HWei:   "5000000000000000000",
				Sales_24H:       5,
				LastSaleAt:      nowProto,
				LastMintAt:      nowProto,
				UpdatedAt:       nowProto,
			},
			expected: &model.CollectionStats{
				CollectionID:    "collection-id",
				TotalItems:      100,
				TotalOwners:     50,
				TotalSales:      25,
				FloorPriceWei:   "1000000000000000000",
				TotalVolumeWei:  "25000000000000000000",
				AveragePriceWei: "1000000000000000000",
				Volume24hWei:    "5000000000000000000",
				Sales24h:        5,
				LastSaleAt:      &now,
				LastMintAt:      &now,
				UpdatedAt:       now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoToGraphQLStats(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("ProtoToGraphQLStats() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("ProtoToGraphQLStats() = nil, want non-nil")
			}

			if result.CollectionID != tt.expected.CollectionID {
				t.Errorf("CollectionID = %v, want %v", result.CollectionID, tt.expected.CollectionID)
			}
			if result.TotalItems != tt.expected.TotalItems {
				t.Errorf("TotalItems = %v, want %v", result.TotalItems, tt.expected.TotalItems)
			}
			if result.TotalOwners != tt.expected.TotalOwners {
				t.Errorf("TotalOwners = %v, want %v", result.TotalOwners, tt.expected.TotalOwners)
			}
			if result.FloorPriceWei != tt.expected.FloorPriceWei {
				t.Errorf("FloorPriceWei = %v, want %v", result.FloorPriceWei, tt.expected.FloorPriceWei)
			}
		})
	}
}

func TestGraphQLCreateInputToProto(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    model.CreateCollectionInput
		userID   string
		expected *pb.CreateCollectionRequest
	}{
		{
			name: "Basic collection input",
			input: model.CreateCollectionInput{
				Name:            "Test Collection",
				Symbol:          "TEST",
				Description:     ptrString("A test collection"),
				TokenStandard:   model.TokenStandardErc721,
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				ChainID:         ptrString("eip155:1"),
				ImageURL:        ptrString("https://example.com/image.png"),
			},
			userID: "user-123",
			expected: &pb.CreateCollectionRequest{
				UserId:          "user-123",
				Name:            "Test Collection",
				Symbol:          "TEST",
				Description:     "A test collection",
				TokenStandard:   pb.TokenStandard_TOKEN_STANDARD_ERC721,
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				ChainId:         "eip155:1",
				ImageUrl:        "https://example.com/image.png",
			},
		},
		{
			name: "Collection with optional fields",
			input: model.CreateCollectionInput{
				Name:               "Full Collection",
				Symbol:             "FULL",
				TokenStandard:      model.TokenStandardErc1155,
				DeployerAddress:    "0x1234567890123456789012345678901234567890",
				RoyaltyFeeBps:      ptrInt(250),
				RoyaltyRecipient:   ptrString("0x1234567890123456789012345678901234567890"),
				MaxSupply:          ptrInt(10000),
				MintPriceAllowlist: ptrString("100000000000000000"),
				MintPricePublic:    ptrString("200000000000000000"),
				MintStartTime:      &now,
				MintLimitPerWallet: ptrInt(5),
			},
			userID: "user-456",
			expected: &pb.CreateCollectionRequest{
				UserId:             "user-456",
				Name:               "Full Collection",
				Symbol:             "FULL",
				TokenStandard:      pb.TokenStandard_TOKEN_STANDARD_ERC1155,
				DeployerAddress:    "0x1234567890123456789012345678901234567890",
				RoyaltyFeeBps:      250,
				RoyaltyRecipient:   "0x1234567890123456789012345678901234567890",
				MaxSupply:          10000,
				MintPriceAllowlist: "100000000000000000",
				MintPricePublic:    "200000000000000000",
				MintStartTime:      timestamppb.New(now),
				MintLimitPerWallet: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GraphQLCreateInputToProto(tt.input, tt.userID)

			if result.UserId != tt.expected.UserId {
				t.Errorf("UserId = %v, want %v", result.UserId, tt.expected.UserId)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
			}
			if result.Symbol != tt.expected.Symbol {
				t.Errorf("Symbol = %v, want %v", result.Symbol, tt.expected.Symbol)
			}
			if result.TokenStandard != tt.expected.TokenStandard {
				t.Errorf("TokenStandard = %v, want %v", result.TokenStandard, tt.expected.TokenStandard)
			}
			if result.DeployerAddress != tt.expected.DeployerAddress {
				t.Errorf("DeployerAddress = %v, want %v", result.DeployerAddress, tt.expected.DeployerAddress)
			}
		})
	}
}

func TestGraphQLUpdateInputToProto(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		userID   string
		input    model.UpdateCollectionInput
		validate func(*testing.T, *pb.UpdateCollectionRequest)
	}{
		{
			name:   "Update description only",
			id:     "collection-123",
			userID: "user-123",
			input: model.UpdateCollectionInput{
				Description: ptrString("Updated description"),
			},
			validate: func(t *testing.T, req *pb.UpdateCollectionRequest) {
				if req.Id != "collection-123" {
					t.Errorf("Id = %v, want collection-123", req.Id)
				}
				if req.UserId != "user-123" {
					t.Errorf("UserId = %v, want user-123", req.UserId)
				}
				if req.Description == nil || *req.Description != "Updated description" {
					t.Errorf("Description = %v, want 'Updated description'", req.Description)
				}
			},
		},
		{
			name:   "Update multiple fields",
			id:     "collection-456",
			userID: "user-456",
			input: model.UpdateCollectionInput{
				Description: ptrString("New description"),
				ImageURL:    ptrString("https://example.com/new.png"),
				WebsiteURL:  ptrString("https://example.com"),
			},
			validate: func(t *testing.T, req *pb.UpdateCollectionRequest) {
				if req.Id != "collection-456" {
					t.Errorf("Id = %v, want collection-456", req.Id)
				}
				if req.Description == nil || *req.Description != "New description" {
					t.Errorf("Description = %v, want 'New description'", req.Description)
				}
				if req.ImageUrl == nil || *req.ImageUrl != "https://example.com/new.png" {
					t.Errorf("ImageUrl = %v, want 'https://example.com/new.png'", req.ImageUrl)
				}
				if req.WebsiteUrl == nil || *req.WebsiteUrl != "https://example.com" {
					t.Errorf("WebsiteUrl = %v, want 'https://example.com'", req.WebsiteUrl)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GraphQLUpdateInputToProto(tt.id, tt.userID, tt.input)
			tt.validate(t, result)
		})
	}
}

// Helper functions
func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}

func ptrCollectionStatus(s model.CollectionStatus) *model.CollectionStatus {
	return &s
}
