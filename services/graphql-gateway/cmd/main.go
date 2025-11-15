package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	pb "github.com/quangdang46/NFT-Marketplace/shared/proto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	authClient   pb.AuthServiceClient
	userClient   pb.UserServiceClient
	walletClient pb.WalletServiceClient
}

func main() {
	log.Println("Starting GraphQL Gateway...")

	// Load config
	httpAddr := getEnv("GATEWAY_HTTP_ADDR", ":8081")
	authURL := getEnv("AUTH_SERVICE_URL", "localhost:50051")
	userURL := getEnv("USER_SERVICE_URL", "localhost:50052")
	walletURL := getEnv("WALLET_SERVICE_URL", "localhost:50053")

	// Connect to services
	authConn, _ := grpc.Dial(authURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	userConn, _ := grpc.Dial(userURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	walletConn, _ := grpc.Dial(walletURL, grpc.WithTransportCredentials(insecure.NewCredentials()))

	server := &Server{
		authClient:   pb.NewAuthServiceClient(authConn),
		userClient:   pb.NewUserServiceClient(userConn),
		walletClient: pb.NewWalletServiceClient(walletConn),
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"GraphQL Gateway - TODO: Implement GraphQL schema"}`)
	})

	log.Printf("GraphQL Gateway listening on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
