package graph

import (
	"context"
	"errors"
	"testing"

	"github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/graph/model"
	"github.com/zunokit/zuno-marketplace-api/services/graphql-gateway/internal/middleware"
	pb "github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mock Collection Service Client
type mockCollectionServiceClient struct {
	pb.CollectionServiceClient
	getCollectionFunc         func(ctx context.Context, in *pb.GetCollectionRequest, opts ...grpc.CallOption) (*pb.GetCollectionResponse, error)
	listCollectionsByUserFunc func(ctx context.Context, in *pb.ListCollectionsByUserRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error)
	listCollectionsFunc       func(ctx context.Context, in *pb.ListCollectionsRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error)
	createCollectionFunc      func(ctx context.Context, in *pb.CreateCollectionRequest, opts ...grpc.CallOption) (*pb.CreateCollectionResponse, error)
	updateCollectionFunc      func(ctx context.Context, in *pb.UpdateCollectionRequest, opts ...grpc.CallOption) (*pb.UpdateCollectionResponse, error)
	addToAllowlistFunc        func(ctx context.Context, in *pb.AddToAllowlistRequest, opts ...grpc.CallOption) (*pb.AddToAllowlistResponse, error)
	deleteCollectionFunc      func(ctx context.Context, in *pb.DeleteCollectionRequest, opts ...grpc.CallOption) (*pb.DeleteCollectionResponse, error)
}

func (m *mockCollectionServiceClient) GetCollection(ctx context.Context, in *pb.GetCollectionRequest, opts ...grpc.CallOption) (*pb.GetCollectionResponse, error) {
	if m.getCollectionFunc != nil {
		return m.getCollectionFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCollectionServiceClient) ListCollectionsByUser(ctx context.Context, in *pb.ListCollectionsByUserRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error) {
	if m.listCollectionsByUserFunc != nil {
		return m.listCollectionsByUserFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCollectionServiceClient) ListCollections(ctx context.Context, in *pb.ListCollectionsRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error) {
	if m.listCollectionsFunc != nil {
		return m.listCollectionsFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCollectionServiceClient) CreateCollection(ctx context.Context, in *pb.CreateCollectionRequest, opts ...grpc.CallOption) (*pb.CreateCollectionResponse, error) {
	if m.createCollectionFunc != nil {
		return m.createCollectionFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCollectionServiceClient) UpdateCollection(ctx context.Context, in *pb.UpdateCollectionRequest, opts ...grpc.CallOption) (*pb.UpdateCollectionResponse, error) {
	if m.updateCollectionFunc != nil {
		return m.updateCollectionFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCollectionServiceClient) AddToAllowlist(ctx context.Context, in *pb.AddToAllowlistRequest, opts ...grpc.CallOption) (*pb.AddToAllowlistResponse, error) {
	if m.addToAllowlistFunc != nil {
		return m.addToAllowlistFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCollectionServiceClient) DeleteCollection(ctx context.Context, in *pb.DeleteCollectionRequest, opts ...grpc.CallOption) (*pb.DeleteCollectionResponse, error) {
	if m.deleteCollectionFunc != nil {
		return m.deleteCollectionFunc(ctx, in, opts...)
	}
	return nil, errors.New("not implemented")
}

// Tests

func TestQueryCollection(t *testing.T) {
	tests := []struct {
		name         string
		collectionID *string
		mockFunc     func(ctx context.Context, in *pb.GetCollectionRequest, opts ...grpc.CallOption) (*pb.GetCollectionResponse, error)
		wantErr      bool
		wantNil      bool
	}{
		{
			name:         "Successfully get collection",
			collectionID: ptrString("test-id"),
			mockFunc: func(ctx context.Context, in *pb.GetCollectionRequest, opts ...grpc.CallOption) (*pb.GetCollectionResponse, error) {
				return &pb.GetCollectionResponse{
					Collection: &pb.Collection{
						Id:              "test-id",
						UserId:          "user-123",
						Name:            "Test Collection",
						Symbol:          "TEST",
						TokenStandard:   pb.TokenStandard_TOKEN_STANDARD_ERC721,
						DeployerAddress: "0x123",
						Status:          pb.CollectionStatus_COLLECTION_STATUS_PENDING,
						TotalSupply:     100,
						TotalMinted:     0,
						IsVerified:      false,
						IsHidden:        false,
						CreatedAt:       timestamppb.Now(),
						UpdatedAt:       timestamppb.Now(),
					},
				}, nil
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:         "Collection not found returns nil",
			collectionID: ptrString("non-existent"),
			mockFunc: func(ctx context.Context, in *pb.GetCollectionRequest, opts ...grpc.CallOption) (*pb.GetCollectionResponse, error) {
				return nil, status.Error(codes.NotFound, "collection not found")
			},
			wantErr: false,
			wantNil: true,
		},
		{
			name:         "gRPC error returns error",
			collectionID: ptrString("error-id"),
			mockFunc: func(ctx context.Context, in *pb.GetCollectionRequest, opts ...grpc.CallOption) (*pb.GetCollectionResponse, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
			wantErr: true,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCollectionServiceClient{
				getCollectionFunc: tt.mockFunc,
			}

			resolver := &Resolver{
				CollectionClient: mockClient,
			}

			queryResolver := &queryResolver{resolver}
			result, err := queryResolver.Collection(context.Background(), tt.collectionID, nil, nil, nil)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.wantNil && result != nil {
				t.Error("Expected nil result, got non-nil")
			}
			if !tt.wantNil && !tt.wantErr && result == nil {
				t.Error("Expected non-nil result, got nil")
			}
		})
	}
}

func TestQueryMyCollections(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		page      *int
		limit     *int
		mockFunc  func(ctx context.Context, in *pb.ListCollectionsByUserRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "Successfully get user collections",
			userID: "user-123",
			page:   ptrInt(1),
			limit:  ptrInt(10),
			mockFunc: func(ctx context.Context, in *pb.ListCollectionsByUserRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error) {
				return &pb.ListCollectionsResponse{
					Collections: []*pb.Collection{
						{
							Id:              "collection-1",
							UserId:          "user-123",
							Name:            "Collection 1",
							Symbol:          "COL1",
							TokenStandard:   pb.TokenStandard_TOKEN_STANDARD_ERC721,
							DeployerAddress: "0x123",
							Status:          pb.CollectionStatus_COLLECTION_STATUS_PENDING,
							TotalSupply:     100,
							TotalMinted:     0,
							IsVerified:      false,
							IsHidden:        false,
							CreatedAt:       timestamppb.Now(),
							UpdatedAt:       timestamppb.Now(),
						},
						{
							Id:              "collection-2",
							UserId:          "user-123",
							Name:            "Collection 2",
							Symbol:          "COL2",
							TokenStandard:   pb.TokenStandard_TOKEN_STANDARD_ERC1155,
							DeployerAddress: "0x456",
							Status:          pb.CollectionStatus_COLLECTION_STATUS_DEPLOYED,
							TotalSupply:     1000,
							TotalMinted:     50,
							IsVerified:      true,
							IsHidden:        false,
							CreatedAt:       timestamppb.Now(),
							UpdatedAt:       timestamppb.Now(),
						},
					},
					TotalCount: 2,
					Page:       1,
					Limit:      10,
				}, nil
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:   "Unauthorized - no user ID",
			userID: "",
			page:   ptrInt(1),
			limit:  ptrInt(10),
			mockFunc: func(ctx context.Context, in *pb.ListCollectionsByUserRequest, opts ...grpc.CallOption) (*pb.ListCollectionsResponse, error) {
				return nil, nil
			},
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCollectionServiceClient{
				listCollectionsByUserFunc: tt.mockFunc,
			}

			resolver := &Resolver{
				CollectionClient: mockClient,
			}

			// Create context with user ID (if provided)
			ctx := context.Background()
			if tt.userID != "" {
				ctx = context.WithValue(ctx, middleware.UserClaimsKey, &middleware.UserClaims{
					UserID: tt.userID,
				})
			}

			queryResolver := &queryResolver{resolver}
			result, err := queryResolver.MyCollections(ctx, tt.page, tt.limit)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && result != nil {
				if len(result.Items) != tt.wantCount {
					t.Errorf("Expected %d collections, got %d", tt.wantCount, len(result.Items))
				}
			}
		})
	}
}

func TestMutationCreateCollection(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		input    model.CreateCollectionInput
		mockFunc func(ctx context.Context, in *pb.CreateCollectionRequest, opts ...grpc.CallOption) (*pb.CreateCollectionResponse, error)
		wantErr  bool
	}{
		{
			name:   "Successfully create collection",
			userID: "user-123",
			input: model.CreateCollectionInput{
				Name:            "New Collection",
				Symbol:          "NEW",
				TokenStandard:   model.TokenStandardErc721,
				DeployerAddress: "0x123",
			},
			mockFunc: func(ctx context.Context, in *pb.CreateCollectionRequest, opts ...grpc.CallOption) (*pb.CreateCollectionResponse, error) {
				return &pb.CreateCollectionResponse{
					Collection: &pb.Collection{
						Id:              "new-collection-id",
						UserId:          in.UserId,
						Name:            in.Name,
						Symbol:          in.Symbol,
						TokenStandard:   in.TokenStandard,
						DeployerAddress: in.DeployerAddress,
						Status:          pb.CollectionStatus_COLLECTION_STATUS_PENDING,
						TotalSupply:     0,
						TotalMinted:     0,
						IsVerified:      false,
						IsHidden:        false,
						CreatedAt:       timestamppb.Now(),
						UpdatedAt:       timestamppb.Now(),
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:   "Unauthorized - no user ID",
			userID: "",
			input: model.CreateCollectionInput{
				Name:            "New Collection",
				Symbol:          "NEW",
				TokenStandard:   model.TokenStandardErc721,
				DeployerAddress: "0x123",
			},
			mockFunc: func(ctx context.Context, in *pb.CreateCollectionRequest, opts ...grpc.CallOption) (*pb.CreateCollectionResponse, error) {
				return nil, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCollectionServiceClient{
				createCollectionFunc: tt.mockFunc,
			}

			resolver := &Resolver{
				CollectionClient: mockClient,
			}

			// Create context with user ID (if provided)
			ctx := context.Background()
			if tt.userID != "" {
				ctx = context.WithValue(ctx, middleware.UserClaimsKey, &middleware.UserClaims{
					UserID: tt.userID,
				})
			}

			mutationResolver := &mutationResolver{resolver}
			result, err := mutationResolver.CreateCollection(ctx, tt.input)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && result != nil {
				if result.Name != tt.input.Name {
					t.Errorf("Expected name %s, got %s", tt.input.Name, result.Name)
				}
				if result.Symbol != tt.input.Symbol {
					t.Errorf("Expected symbol %s, got %s", tt.input.Symbol, result.Symbol)
				}
			}
		})
	}
}

func TestMutationUpdateCollection(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		id       string
		input    model.UpdateCollectionInput
		mockFunc func(ctx context.Context, in *pb.UpdateCollectionRequest, opts ...grpc.CallOption) (*pb.UpdateCollectionResponse, error)
		wantErr  bool
	}{
		{
			name:   "Successfully update collection",
			userID: "user-123",
			id:     "collection-123",
			input: model.UpdateCollectionInput{
				Description: ptrString("Updated description"),
			},
			mockFunc: func(ctx context.Context, in *pb.UpdateCollectionRequest, opts ...grpc.CallOption) (*pb.UpdateCollectionResponse, error) {
				return &pb.UpdateCollectionResponse{
					Collection: &pb.Collection{
						Id:              in.Id,
						UserId:          in.UserId,
						Name:            "Test Collection",
						Symbol:          "TEST",
						Description:     *in.Description,
						TokenStandard:   pb.TokenStandard_TOKEN_STANDARD_ERC721,
						DeployerAddress: "0x123",
						Status:          pb.CollectionStatus_COLLECTION_STATUS_PENDING,
						TotalSupply:     100,
						TotalMinted:     0,
						IsVerified:      false,
						IsHidden:        false,
						CreatedAt:       timestamppb.Now(),
						UpdatedAt:       timestamppb.Now(),
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:   "Unauthorized - no user ID",
			userID: "",
			id:     "collection-123",
			input: model.UpdateCollectionInput{
				Description: ptrString("Updated description"),
			},
			mockFunc: func(ctx context.Context, in *pb.UpdateCollectionRequest, opts ...grpc.CallOption) (*pb.UpdateCollectionResponse, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:   "Permission denied - not owner",
			userID: "user-123",
			id:     "collection-456",
			input: model.UpdateCollectionInput{
				Description: ptrString("Updated description"),
			},
			mockFunc: func(ctx context.Context, in *pb.UpdateCollectionRequest, opts ...grpc.CallOption) (*pb.UpdateCollectionResponse, error) {
				return nil, status.Error(codes.PermissionDenied, "you don't own this collection")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCollectionServiceClient{
				updateCollectionFunc: tt.mockFunc,
			}

			resolver := &Resolver{
				CollectionClient: mockClient,
			}

			// Create context with user ID (if provided)
			ctx := context.Background()
			if tt.userID != "" {
				ctx = context.WithValue(ctx, middleware.UserClaimsKey, &middleware.UserClaims{
					UserID: tt.userID,
				})
			}

			mutationResolver := &mutationResolver{resolver}
			result, err := mutationResolver.UpdateCollection(ctx, tt.id, tt.input)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && result != nil {
				if result.ID != tt.id {
					t.Errorf("Expected ID %s, got %s", tt.id, result.ID)
				}
			}
		})
	}
}

func TestMutationDeleteCollection(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		id       string
		mockFunc func(ctx context.Context, in *pb.DeleteCollectionRequest, opts ...grpc.CallOption) (*pb.DeleteCollectionResponse, error)
		wantErr  bool
		wantBool bool
	}{
		{
			name:   "Successfully delete collection",
			userID: "user-123",
			id:     "collection-123",
			mockFunc: func(ctx context.Context, in *pb.DeleteCollectionRequest, opts ...grpc.CallOption) (*pb.DeleteCollectionResponse, error) {
				return &pb.DeleteCollectionResponse{
					Success: true,
				}, nil
			},
			wantErr:  false,
			wantBool: true,
		},
		{
			name:   "Unauthorized - no user ID",
			userID: "",
			id:     "collection-123",
			mockFunc: func(ctx context.Context, in *pb.DeleteCollectionRequest, opts ...grpc.CallOption) (*pb.DeleteCollectionResponse, error) {
				return nil, nil
			},
			wantErr:  true,
			wantBool: false,
		},
		{
			name:   "Permission denied - not owner",
			userID: "user-123",
			id:     "collection-456",
			mockFunc: func(ctx context.Context, in *pb.DeleteCollectionRequest, opts ...grpc.CallOption) (*pb.DeleteCollectionResponse, error) {
				return nil, status.Error(codes.PermissionDenied, "you don't own this collection")
			},
			wantErr:  true,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockCollectionServiceClient{
				deleteCollectionFunc: tt.mockFunc,
			}

			resolver := &Resolver{
				CollectionClient: mockClient,
			}

			// Create context with user ID (if provided)
			ctx := context.Background()
			if tt.userID != "" {
				ctx = context.WithValue(ctx, middleware.UserClaimsKey, &middleware.UserClaims{
					UserID: tt.userID,
				})
			}

			mutationResolver := &mutationResolver{resolver}
			result, err := mutationResolver.DeleteCollection(ctx, tt.id)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.wantBool {
				t.Errorf("Expected %v, got %v", tt.wantBool, result)
			}
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
