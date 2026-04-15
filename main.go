package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gagliardetto/solana-go"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

func main() {
	var endpoint, token, commitment string

	flag.StringVar(&endpoint, "endpoint", "http://localhost:10000", "grpc server url")
	flag.StringVar(&token, "token", "", "auth token")
	flag.StringVar(&commitment, "commitment", "PROCRESSED", "commitment level")

	flag.Parse()

	fmt.Printf("ENDPOINT: %s TOKEN: %s CommitmentLevel: %s", endpoint, token, commitment)

	if token == "" {
		fmt.Println("ERR: token is not set")
		return
	}

	subscribeTx(endpoint, token, getCommitmentLevel(commitment))
}

func getCommitmentLevel(commitment string) *pb.CommitmentLevel {
	result := pb.CommitmentLevel_PROCESSED
	switch commitment {
	case "FINALIZED":
		result = pb.CommitmentLevel_CONFIRMED
		return &result
	case "CONFIRMED":
		result = pb.CommitmentLevel_PROCESSED
		return &result
	}
	return &result
}

// func subscribeBlck(endpoint, token string) {
// }

func subscribeTx(endpoint string, token string, commitment *pb.CommitmentLevel) {
	builder, err := yellowstone.BuildFromShared(endpoint)
	if err != nil {
		log.Fatalf("Error building client: %v", err)
	}

	clientBuilder := builder.XToken(token).KeepAliveWhileIdle(true)
	grpcClient, err := clientBuilder.Connect(context.Background())
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	defer grpcClient.Close()
	failed := false
	req := &pb.SubscribeRequest{
		Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
			"tx_filter": {
				AccountInclude: []string{
					solana.TokenProgramID.String(),
				},
				Failed: &failed,
			},
		},
		Commitment: commitment,
	}

	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("💸 Listening for transaction updates...")

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Transaction:
			txUpdate := update.GetTransaction()
			tx := txUpdate.Transaction

			if len(tx.Meta.GetPostTokenBalances()) == 0 {
				return nil
			}

			arrived := time.Now()

			fmt.Printf("\n💸 Transaction Update:\n")
			fmt.Printf("   Signature: %s\n", solana.SignatureFromBytes(tx.Signature).String())
			fmt.Printf("   Arrived time: %d\n", arrived.UnixMilli())
			fmt.Printf("   Sent at: %d\n", update.CreatedAt.AsTime().UnixMilli())
			fmt.Printf("   Delay : %s", arrived.Sub(update.CreatedAt.AsTime()).String())
			fmt.Printf("   Slot: %d\n", txUpdate.Slot)

		case *pb.SubscribeUpdate_Ping:
			return nil

		case *pb.SubscribeUpdate_Pong:
			return nil

		default:
			fmt.Printf("⚠️  Unexpected update type: %T\n", update.GetUpdateOneof())
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting client: %v", err)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	grpcClient.Close()
	fmt.Println("✅ Transaction subscription example completed")
}
