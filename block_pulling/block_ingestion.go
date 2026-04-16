package main

import (
	"context"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

type BlockIngestionService struct {
	cli *yellowstone.GeyserGrpcClient

	dec *Decoder
}

func NewBlockIngestionService(dec *Decoder) *BlockIngestionService {
	return &BlockIngestionService{
		dec: dec,
	}
}

func (svc *BlockIngestionService) Close() {
	if svc.cli != nil {
		svc.cli.Close()
	}
}

func (svc *BlockIngestionService) Subscribe(endpoint, token string, commitment *pb.CommitmentLevel) {
	builder, err := yellowstone.BuildFromShared(endpoint)
	if err != nil {
		log.Fatalf("Error building client: %v", err)
	}

	clientBuilder := builder.XToken(token).KeepAliveWhileIdle(true).MaxDecodingMessageSize(100 * 1024 * 1024)
	grpcClient, err := clientBuilder.Connect(context.Background())
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	svc.cli = grpcClient

	includeTx := true
	includeAccounts := false
	includeEntries := false

	req := &pb.SubscribeRequest{
		Blocks: map[string]*pb.SubscribeRequestFilterBlocks{
			"block_filter": {
				IncludeTransactions: &includeTx,
				IncludeAccounts:     &includeAccounts,
				IncludeEntries:      &includeEntries,
			},
		},
		Commitment: commitment,
	}
	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("Listening for block updates...")
	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Block:
			block := update.GetBlock()
			fmt.Println("BlockTime: ", block.BlockTime.Timestamp, "Slot", block.Slot)
			for _, tx := range block.GetTransactions() {
				if tx == nil {
					continue
				}

				arrived := time.Now()

				txCtx := TxContext{
					GeyserSentTime:     update.CreatedAt.AsTime(),
					ServerReceivedTime: arrived,
					Slot:               block.Slot,
					BlockTx:            tx,
					BlockTime:          block.BlockTime.Timestamp,
				}
				svc.dec.Queue(&txCtx)
			}
		default:
			return nil
		}
		return nil
	})
}
