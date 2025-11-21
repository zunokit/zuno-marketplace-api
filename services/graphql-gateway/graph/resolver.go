package graph

import (
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AuthClient       pb.AuthServiceClient
	UserClient       pb.UserServiceClient
	WalletClient     pb.WalletServiceClient
	CollectionClient pb.CollectionServiceClient
}
