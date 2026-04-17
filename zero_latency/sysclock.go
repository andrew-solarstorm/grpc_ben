package main

import (
	"context"
	"log"

	"github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	lru "github.com/hashicorp/golang-lru/v2"
)

// ref: https://docs.rs/solana-program/latest/solana_program/sysvar/clock/struct.Clock.html
type SysVarClock struct {
	Slot                uint64
	EpochStartTimestamp int64
	Epoch               uint64
	LeaderScheduleEpoch uint64
	UnixTimestamp       int64
}

func decodeClock(data []byte) (*SysVarClock, error) {
	var clock SysVarClock

	decoder := bin.NewBinDecoder(data)

	err := decoder.Decode(&clock)
	if err != nil {
		return nil, err
	}

	return &clock, nil
}

type SystemClock struct {
	cli   *yellowstone.GeyserGrpcClient
	slots *lru.Cache[uint64, int64]
}

func NewSysClock() *SystemClock {
	cache, _ := lru.New[uint64, int64](30)
	return &SystemClock{
		slots: cache,
	}
}

func (svc *SystemClock) TimeStamp(slot uint64) int64 {
	slotTime, _ := svc.slots.Get(slot)
	return slotTime
}

func (svc *SystemClock) Close() {
	if svc.cli != nil {
		svc.cli.Close()
	}
}

func (svc *SystemClock) subscribe(endpoint, token string, commitment *pb.CommitmentLevel) {
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

	req := &pb.SubscribeRequest{
		Accounts: map[string]*pb.SubscribeRequestFilterAccounts{
			"sysvar_filter": {
				Account: []string{
					solana.SysVarClockPubkey.String(),
				},
			},
		},
	}

	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Account:
			accountUpdate := update.GetAccount()
			account := accountUpdate.Account

			clock, err := decodeClock(account.Data)
			if err != nil {
				return nil
			}

			if _, ok := svc.slots.Get(clock.Slot); !ok {
				svc.slots.Add(clock.Slot, clock.UnixTimestamp)
			}

		default:
			return nil
		}
		return nil
	})
}
