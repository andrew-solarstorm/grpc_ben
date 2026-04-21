package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

func getCommitmentLevel(commitment string) *pb.CommitmentLevel {
	result := pb.CommitmentLevel_PROCESSED
	switch commitment {
	case "FINALIZED":
		result = pb.CommitmentLevel_FINALIZED
		return &result
	case "CONFIRMED":
		result = pb.CommitmentLevel_CONFIRMED
		return &result
	}
	return &result
}

func main() {
	var endpoint, token, commitmentStr string
	var interUpd bool

	flag.StringVar(&endpoint, "endpoint", "http://localhost:10000", "grpc server url")
	flag.StringVar(&token, "token", "", "auth token")
	flag.StringVar(&commitmentStr, "commitment", "PROCESSED", "commitment level")
	flag.BoolVar(&interUpd, "interUpd", true, "enable interslot updates")

	flag.Parse()

	fmt.Printf("ENDPOINT: %s TOKEN: %s CommitmentLevel: %s InterUpd: %t\n", endpoint, token, commitmentStr, interUpd)

	subscribe(endpoint, token, getCommitmentLevel(commitmentStr), interUpd)
}

func subscribe(endpoint string, token string, commitment *pb.CommitmentLevel, interUpd bool) {
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
	req := &pb.SubscribeRequest{
		Slots: map[string]*pb.SubscribeRequestFilterSlots{
			"slot": {
				InterslotUpdates: &interUpd,
			},
		},
	}

	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("📦 Listening for slot updates...")

	go func() {
		err = grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
			switch update.GetUpdateOneof().(type) {
			case *pb.SubscribeUpdate_Slot:
				slot := update.GetSlot()
				fmt.Printf("📦 Slot Update:\n")
				fmt.Printf("   Slot: %d\n", slot.Slot)
				fmt.Printf("   Status: %s\n", slot.Status)
				fmt.Printf("   Created: %s\n", update.CreatedAt)

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
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	grpcClient.Close()
	fmt.Println("✅ Slot subscription example completed")
}
