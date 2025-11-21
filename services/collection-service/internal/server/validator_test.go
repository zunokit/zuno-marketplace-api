package server

import (
	"testing"

	"github.com/google/uuid"
	"github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		fieldName string
		wantErr   bool
		errCode   codes.Code
	}{
		{
			name:      "valid UUID",
			id:        "550e8400-e29b-41d4-a716-446655440000",
			fieldName: "user_id",
			wantErr:   false,
		},
		{
			name:      "empty UUID",
			id:        "",
			fieldName: "user_id",
			wantErr:   true,
			errCode:   codes.InvalidArgument,
		},
		{
			name:      "invalid UUID format",
			id:        "not-a-uuid",
			fieldName: "collection_id",
			wantErr:   true,
			errCode:   codes.InvalidArgument,
		},
		{
			name:      "UUID with wrong length",
			id:        "550e8400-e29b-41d4",
			fieldName: "id",
			wantErr:   true,
			errCode:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateUUID(tt.id, tt.fieldName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateUUID() error = nil, wantErr %v", tt.wantErr)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("ValidateUUID() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("ValidateUUID() error code = %v, want %v", st.Code(), tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateUUID() unexpected error = %v", err)
					return
				}
				if got == uuid.Nil {
					t.Errorf("ValidateUUID() returned nil UUID")
				}
			}
		})
	}
}

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name      string
		page      int32
		limit     int32
		maxLimit  int
		wantPage  int
		wantLimit int
	}{
		{
			name:      "valid pagination",
			page:      2,
			limit:     20,
			maxLimit:  100,
			wantPage:  2,
			wantLimit: 20,
		},
		{
			name:      "page less than 1 defaults to 1",
			page:      0,
			limit:     10,
			maxLimit:  0,
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:      "negative page defaults to 1",
			page:      -5,
			limit:     10,
			maxLimit:  0,
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:      "limit less than 1 defaults to 10",
			page:      1,
			limit:     0,
			maxLimit:  0,
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:      "limit exceeds max is capped",
			page:      1,
			limit:     200,
			maxLimit:  100,
			wantPage:  1,
			wantLimit: 100,
		},
		{
			name:      "no max limit allows any limit",
			page:      1,
			limit:     500,
			maxLimit:  0,
			wantPage:  1,
			wantLimit: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePaginationParams(tt.page, tt.limit, tt.maxLimit)

			if got.Page != tt.wantPage {
				t.Errorf("ValidatePaginationParams().Page = %v, want %v", got.Page, tt.wantPage)
			}
			if got.Limit != tt.wantLimit {
				t.Errorf("ValidatePaginationParams().Limit = %v, want %v", got.Limit, tt.wantLimit)
			}
		})
	}
}

func TestValidateCreateCollectionRequest(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *pb.CreateCollectionRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid request",
			req: &pb.CreateCollectionRequest{
				UserId:          validUUID,
				Name:            "Test Collection",
				Symbol:          "TEST",
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				ImageUrl:        "https://example.com/image.png",
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			req: &pb.CreateCollectionRequest{
				Name:            "Test Collection",
				Symbol:          "TEST",
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				ImageUrl:        "https://example.com/image.png",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing name",
			req: &pb.CreateCollectionRequest{
				UserId:          validUUID,
				Symbol:          "TEST",
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				ImageUrl:        "https://example.com/image.png",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing symbol",
			req: &pb.CreateCollectionRequest{
				UserId:          validUUID,
				Name:            "Test Collection",
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				ImageUrl:        "https://example.com/image.png",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing deployer_address",
			req: &pb.CreateCollectionRequest{
				UserId:   validUUID,
				Name:     "Test Collection",
				Symbol:   "TEST",
				ImageUrl: "https://example.com/image.png",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing image_url",
			req: &pb.CreateCollectionRequest{
				UserId:          validUUID,
				Name:            "Test Collection",
				Symbol:          "TEST",
				DeployerAddress: "0x1234567890123456789012345678901234567890",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.CreateCollectionRequest{
				UserId:          "invalid-uuid",
				Name:            "Test Collection",
				Symbol:          "TEST",
				DeployerAddress: "0x1234567890123456789012345678901234567890",
				ImageUrl:        "https://example.com/image.png",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateCollectionRequest(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateCreateCollectionRequest() error = nil, wantErr %v", tt.wantErr)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("ValidateCreateCollectionRequest() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("ValidateCreateCollectionRequest() error code = %v, want %v", st.Code(), tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCreateCollectionRequest() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateGetCollectionRequest(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *pb.GetCollectionRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid request with ID",
			req: &pb.GetCollectionRequest{
				Id: validUUID,
			},
			wantErr: false,
		},
		{
			name: "valid request with contract address and chain ID",
			req: &pb.GetCollectionRequest{
				ContractAddress: "0x1234567890123456789012345678901234567890",
				ChainId:         "eip155:1",
			},
			wantErr: false,
		},
		{
			name:    "missing both ID and contract info",
			req:     &pb.GetCollectionRequest{},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "contract address without chain ID",
			req: &pb.GetCollectionRequest{
				ContractAddress: "0x1234567890123456789012345678901234567890",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "chain ID without contract address",
			req: &pb.GetCollectionRequest{
				ChainId: "eip155:1",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid UUID format",
			req: &pb.GetCollectionRequest{
				Id: "invalid-uuid",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetCollectionRequest(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateGetCollectionRequest() error = nil, wantErr %v", tt.wantErr)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("ValidateGetCollectionRequest() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("ValidateGetCollectionRequest() error code = %v, want %v", st.Code(), tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateGetCollectionRequest() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateUpdateCollectionRequest(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *pb.UpdateCollectionRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid request",
			req: &pb.UpdateCollectionRequest{
				Id:     validUUID,
				UserId: validUUID,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			req: &pb.UpdateCollectionRequest{
				UserId: validUUID,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing user_id",
			req: &pb.UpdateCollectionRequest{
				Id: validUUID,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid ID format",
			req: &pb.UpdateCollectionRequest{
				Id:     "invalid-uuid",
				UserId: validUUID,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.UpdateCollectionRequest{
				Id:     validUUID,
				UserId: "invalid-uuid",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateCollectionRequest(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateUpdateCollectionRequest() error = nil, wantErr %v", tt.wantErr)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("ValidateUpdateCollectionRequest() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("ValidateUpdateCollectionRequest() error code = %v, want %v", st.Code(), tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateUpdateCollectionRequest() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateDeleteCollectionRequest(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *pb.DeleteCollectionRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid request",
			req: &pb.DeleteCollectionRequest{
				Id:     validUUID,
				UserId: validUUID,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			req: &pb.DeleteCollectionRequest{
				UserId: validUUID,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing user_id",
			req: &pb.DeleteCollectionRequest{
				Id: validUUID,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid ID format",
			req: &pb.DeleteCollectionRequest{
				Id:     "invalid-uuid",
				UserId: validUUID,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.DeleteCollectionRequest{
				Id:     validUUID,
				UserId: "invalid-uuid",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeleteCollectionRequest(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateDeleteCollectionRequest() error = nil, wantErr %v", tt.wantErr)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("ValidateDeleteCollectionRequest() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("ValidateDeleteCollectionRequest() error code = %v, want %v", st.Code(), tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDeleteCollectionRequest() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateAddToAllowlistRequest(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *pb.AddToAllowlistRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid request",
			req: &pb.AddToAllowlistRequest{
				CollectionId:    validUUID,
				UserId:          validUUID,
				WalletAddresses: []string{"0x1234567890123456789012345678901234567890"},
			},
			wantErr: false,
		},
		{
			name: "missing collection_id",
			req: &pb.AddToAllowlistRequest{
				UserId:          validUUID,
				WalletAddresses: []string{"0x1234567890123456789012345678901234567890"},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing user_id",
			req: &pb.AddToAllowlistRequest{
				CollectionId:    validUUID,
				WalletAddresses: []string{"0x1234567890123456789012345678901234567890"},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "empty wallet addresses",
			req: &pb.AddToAllowlistRequest{
				CollectionId:    validUUID,
				UserId:          validUUID,
				WalletAddresses: []string{},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid collection_id format",
			req: &pb.AddToAllowlistRequest{
				CollectionId:    "invalid-uuid",
				UserId:          validUUID,
				WalletAddresses: []string{"0x1234567890123456789012345678901234567890"},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.AddToAllowlistRequest{
				CollectionId:    validUUID,
				UserId:          "invalid-uuid",
				WalletAddresses: []string{"0x1234567890123456789012345678901234567890"},
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddToAllowlistRequest(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAddToAllowlistRequest() error = nil, wantErr %v", tt.wantErr)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("ValidateAddToAllowlistRequest() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("ValidateAddToAllowlistRequest() error code = %v, want %v", st.Code(), tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAddToAllowlistRequest() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateListCollectionsByUserRequest(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *pb.ListCollectionsByUserRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "valid request",
			req: &pb.ListCollectionsByUserRequest{
				UserId: validUUID,
			},
			wantErr: false,
		},
		{
			name:    "missing user_id",
			req:     &pb.ListCollectionsByUserRequest{},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid user_id format",
			req: &pb.ListCollectionsByUserRequest{
				UserId: "invalid-uuid",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListCollectionsByUserRequest(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateListCollectionsByUserRequest() error = nil, wantErr %v", tt.wantErr)
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("ValidateListCollectionsByUserRequest() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.errCode {
					t.Errorf("ValidateListCollectionsByUserRequest() error code = %v, want %v", st.Code(), tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateListCollectionsByUserRequest() unexpected error = %v", err)
				}
			}
		})
	}
}
