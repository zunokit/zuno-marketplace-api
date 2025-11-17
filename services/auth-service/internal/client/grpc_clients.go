package client

import (
	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceClients struct {
	UserClient   pb.UserServiceClient
	WalletClient pb.WalletServiceClient
}

func NewServiceClients(userURL, walletURL string) (*ServiceClients, error) {
	// Connect to user service
	userConn, err := grpc.Dial(userURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// Connect to wallet service
	walletConn, err := grpc.Dial(walletURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &ServiceClients{
		UserClient:   pb.NewUserServiceClient(userConn),
		WalletClient: pb.NewWalletServiceClient(walletConn),
	}, nil
}
