package server

import (
	"github.com/google/uuid"
	"github.com/zunokit/zuno-marketplace-api/shared/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidateUUID validates and parses a UUID string
func ValidateUUID(id, fieldName string) (uuid.UUID, error) {
	if id == "" {
		return uuid.Nil, status.Errorf(codes.InvalidArgument, "%s is required", fieldName)
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, status.Errorf(codes.InvalidArgument, "invalid %s", fieldName)
	}
	return parsed, nil
}

// PaginationParams holds validated pagination parameters
type PaginationParams struct {
	Page  int
	Limit int
}

// ValidatePaginationParams validates and normalizes pagination parameters
func ValidatePaginationParams(page, limit int32, maxLimit int) PaginationParams {
	p := int(page)
	if p < 1 {
		p = 1
	}

	l := int(limit)
	if l < 1 {
		l = 10
	}
	if maxLimit > 0 && l > maxLimit {
		l = maxLimit
	}

	return PaginationParams{
		Page:  p,
		Limit: l,
	}
}

// ValidateCreateCollectionRequest validates the create collection request
func ValidateCreateCollectionRequest(req *pb.CreateCollectionRequest) error {
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Name == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Symbol == "" {
		return status.Error(codes.InvalidArgument, "symbol is required")
	}
	if req.DeployerAddress == "" {
		return status.Error(codes.InvalidArgument, "deployer_address is required")
	}
	if req.ImageUrl == "" {
		return status.Error(codes.InvalidArgument, "image_url is required")
	}

	// Validate user_id format
	if _, err := ValidateUUID(req.UserId, "user_id"); err != nil {
		return err
	}

	return nil
}

// ValidateGetCollectionRequest validates the get collection request
func ValidateGetCollectionRequest(req *pb.GetCollectionRequest) error {
	// Must provide either ID or contract_address+chain_id
	hasID := req.Id != ""
	hasContract := req.ContractAddress != "" && req.ChainId != ""

	if !hasID && !hasContract {
		return status.Error(codes.InvalidArgument, "must provide id or contract_address+chain_id")
	}

	if hasID {
		if _, err := ValidateUUID(req.Id, "id"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateCollectionRequest validates the update collection request
func ValidateUpdateCollectionRequest(req *pb.UpdateCollectionRequest) error {
	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "id is required")
	}
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Validate IDs format
	if _, err := ValidateUUID(req.Id, "id"); err != nil {
		return err
	}
	if _, err := ValidateUUID(req.UserId, "user_id"); err != nil {
		return err
	}

	return nil
}

// ValidateDeleteCollectionRequest validates the delete collection request
func ValidateDeleteCollectionRequest(req *pb.DeleteCollectionRequest) error {
	if req.Id == "" {
		return status.Error(codes.InvalidArgument, "id is required")
	}
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Validate IDs format
	if _, err := ValidateUUID(req.Id, "id"); err != nil {
		return err
	}
	if _, err := ValidateUUID(req.UserId, "user_id"); err != nil {
		return err
	}

	return nil
}

// ValidateAddToAllowlistRequest validates the add to allowlist request
func ValidateAddToAllowlistRequest(req *pb.AddToAllowlistRequest) error {
	if req.CollectionId == "" {
		return status.Error(codes.InvalidArgument, "collection_id is required")
	}
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	if len(req.WalletAddresses) == 0 {
		return status.Error(codes.InvalidArgument, "wallet_addresses cannot be empty")
	}

	// Validate IDs format
	if _, err := ValidateUUID(req.CollectionId, "collection_id"); err != nil {
		return err
	}
	if _, err := ValidateUUID(req.UserId, "user_id"); err != nil {
		return err
	}

	return nil
}

// ValidateListCollectionsByUserRequest validates the list collections by user request
func ValidateListCollectionsByUserRequest(req *pb.ListCollectionsByUserRequest) error {
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Validate user_id format
	if _, err := ValidateUUID(req.UserId, "user_id"); err != nil {
		return err
	}

	return nil
}
