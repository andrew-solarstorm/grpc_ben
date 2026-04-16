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

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

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

func main() {
	var endpoint, token, commitmentStr string

	flag.StringVar(&endpoint, "endpoint", "http://localhost:10000", "grpc server url")
	flag.StringVar(&token, "token", "", "auth token")
	flag.StringVar(&commitmentStr, "commitment", "PROCESSED", "commitment level")

	flag.Parse()

	subscribe(endpoint, token, getCommitmentLevel(commitmentStr))
}

type Clock struct {
	Slot                uint64
	EpochStartTimestamp int64
	Epoch               uint64
	LeaderScheduleEpoch uint64
	UnixTimestamp       int64
}

func decodeClock(data []byte) (*Clock, error) {
	var clock Clock

	// Create a new decoder from the raw bytes
	decoder := bin.NewBinDecoder(data)

	// Decode the bytes into the struct
	err := decoder.Decode(&clock)
	if err != nil {
		return nil, err
	}

	return &clock, nil
}

func subscribe(endpoint string, token string, commitment *pb.CommitmentLevel) {
	builder, err := yellowstone.BuildFromShared(endpoint)
	if err != nil {
		log.Fatalf("Error building client: %v", err)
	}

	clientBuilder := builder.XToken(token).KeepAliveWhileIdle(true)

	// if tlsConfig := getTLSConfig(endpoint); tlsConfig != nil {
	// 	clientBuilder = clientBuilder.TLSConfig(tlsConfig)
	// }

	grpcClient, err := clientBuilder.Connect(context.Background())
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	defer grpcClient.Close()

	req := &pb.SubscribeRequest{
		Accounts: map[string]*pb.SubscribeRequestFilterAccounts{
			"account_filter": {
				Account: []string{
					solana.SysVarClockPubkey.String(),
				},
			},
		},
		BlocksMeta: map[string]*pb.SubscribeRequestFilterBlocksMeta{
			"blck_meta": {},
		},
	}

	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("👤 Listening for account updates...")

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Account:
			accountUpdate := update.GetAccount()
			account := accountUpdate.Account

			// fmt.Printf("\n👤 Account Update:\n")
			// fmt.Printf("   Pubkey: %s\n", solana.PublicKeyFromBytes(account.Pubkey).String())
			// fmt.Printf("   Owner: %s\n", solana.PublicKeyFromBytes(account.Owner).String())
			// fmt.Printf("   Lamports: %d\n", account.Lamports)
			// fmt.Printf("   Executable: %v\n", account.Executable)
			// fmt.Printf("   Rent Epoch: %d\n", account.RentEpoch)
			// fmt.Printf("   Write Version: %d\n", account.WriteVersion)
			// fmt.Printf("   Data Length: %d bytes\n", len(account.Data))
			// fmt.Printf("   Slot: %d\n", accountUpdate.Slot)
			// fmt.Printf("   Is Startup: %v\n", accountUpdate.IsStartup)

			clock, err := decodeClock(account.Data)
			if err != nil {
				return nil
			}

			fmt.Printf("SysVarClock.UnixTimestamp: %d Arrived: %d  UpdateCreated at: %d \n", clock.UnixTimestamp, time.Now().UnixMilli(), update.CreatedAt.AsTime().UnixMicro())

		case *pb.SubscribeUpdate_Slot:
			// slot := update.GetSlot()
			// fmt.Printf("📦 Slot: %d (Status: %s)\n", slot.Slot, slot.Status)

		case *pb.SubscribeUpdate_Ping:
			return nil

		case *pb.SubscribeUpdate_BlockMeta:
			blockMeta := update.GetBlockMeta()
			fmt.Printf("Block Time: %d Arrived: %d  UpdateCreated at: %d\n\n", blockMeta.BlockTime.Timestamp, time.Now().UnixMilli(), update.CreatedAt.AsTime().UnixMicro())

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
	fmt.Println("✅ Account subscription example completed")
}
