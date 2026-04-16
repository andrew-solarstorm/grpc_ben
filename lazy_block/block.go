package main

import (
	"context"
	"log"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	lru "github.com/hashicorp/golang-lru/v2"
)

type BlockService struct {
	block *lru.Cache[uint64, *pb.SubscribeUpdateBlockMeta]
	cli   *yellowstone.GeyserGrpcClient

	evtAggSvc *EventAggregator
}

func NewBlockService() *BlockService {
	cache, _ := lru.New[uint64, *pb.SubscribeUpdateBlockMeta](30)
	return &BlockService{
		block: cache,
	}
}

func (svc *BlockService) getBlockTime(slot uint64) int64 {
	result, ok := svc.block.Get(slot)
	if ok {
		return result.BlockTime.Timestamp
	}
	return 0
}

func (svc *BlockService) Subscribe(endpoint, token string, commitment *pb.CommitmentLevel) {
	builder, err := yellowstone.BuildFromShared(endpoint)
	if err != nil {
		log.Fatalf("Error building client: %v", err)
	}

	clientBuilder := builder.XToken(token).KeepAliveWhileIdle(true).MaxDecodingMessageSize(100 * 1024 * 1024)

	grpcClient, err := clientBuilder.Connect(context.Background())
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
		return
	}

	svc.cli = grpcClient

	req := &pb.SubscribeRequest{
		BlocksMeta: map[string]*pb.SubscribeRequestFilterBlocksMeta{
			"block_meta_filter": {},
		},
	}
	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_BlockMeta:
			blockMeta := update.GetBlockMeta()
			svc.block.Add(blockMeta.Slot, blockMeta)

			if svc.evtAggSvc != nil {
				svc.evtAggSvc.OnBlockMeta(blockMeta.Slot, blockMeta.BlockTime.Timestamp)
			}
		default:
		}
		return nil
	})
}

func (svc *BlockService) Close() {
	if svc.cli != nil {
		svc.cli.Close()
	}
}
