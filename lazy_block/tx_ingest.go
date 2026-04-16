package main

import (
	"context"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"github.com/gagliardetto/solana-go"
)

type TxIngestService struct {
	blckSvc *BlockService
	decSvc  *Decoder
	cli     *yellowstone.GeyserGrpcClient
}

func NewTxIngestService(blckSvc *BlockService, decSvc *Decoder) *TxIngestService {
	return &TxIngestService{
		blckSvc: blckSvc,
		decSvc:  decSvc,
	}
}

func (svc *TxIngestService) Subscribe(endpoint, token string, commitment *pb.CommitmentLevel) {
	builder, err := yellowstone.BuildFromShared(endpoint)
	if err != nil {
		log.Fatalf("Error building client: %v", err)
	}

	clientBuilder := builder.XToken(token).KeepAliveWhileIdle(true)
	grpcClient, err := clientBuilder.Connect(context.Background())
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	svc.cli = grpcClient
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

			txCtx := TxContext{
				GeyserSentTime:     update.CreatedAt.AsTime(),
				ServerReceivedTime: arrived,
				Upd:                txUpdate,
				Slot:               txUpdate.Slot,
			}

			svc.decSvc.Queue(&txCtx)

			// fmt.Printf("\n💸 Transaction Update:\n")
			// fmt.Printf("   Signature: %s\n", solana.SignatureFromBytes(tx.Signature).String())
			// fmt.Printf("   Arrived time: %d\n", arrived.UnixMilli())
			// fmt.Printf("   Sent at: %d\n", update.CreatedAt.AsTime().UnixMilli())
			// fmt.Printf("   Delay : %s \n", arrived.Sub(update.CreatedAt.AsTime()).String())
			// fmt.Printf("   BlockTime: %d \n", svc.blckSvc.getBlockTime(txUpdate.Slot))
			// fmt.Printf("   Slot: %d\n", txUpdate.Slot)

		default:
			return nil
		}
		return nil
	})
}

func (svc *TxIngestService) Close() {
	if svc.cli != nil {
		svc.cli.Close()
	}
}
