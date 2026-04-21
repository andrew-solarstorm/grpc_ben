package main

import (
	"log"
	"sync"
	"time"

	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

const BatchSize int8 = 10

type Buffer struct {
	mu  *sync.RWMutex
	txs []*pb.SubscribeUpdateTransactionInfo

	slot uint64

	clock    *SystemClock
	blockSvc *BlockBuilder
}

func (b *Buffer) Add(slot uint64, tx *pb.SubscribeUpdateTransactionInfo) {
	log.Println("Adding tx to buffer", slot)
	b.mu.Lock()
	defer b.mu.Unlock()
	if slot < b.slot {
		return
	}

	if slot > b.slot || len(b.txs) >= int(BatchSize) {
		temp := b.txs
		go b.buildBlock(b.slot, temp)

		b.txs = make([]*pb.SubscribeUpdateTransactionInfo, 0)
		b.slot = slot
	}
	b.txs = append(b.txs, tx)
}

func (b *Buffer) buildBlock(slot uint64, txs []*pb.SubscribeUpdateTransactionInfo) {
	log.Printf("Building Batch: %d | Len: %d\n", slot, len(txs))
	block := &LocalBlock{
		Slot:         slot,
		Blocktime:    b.clock.TimeStamp(slot),
		CreatedAt:    time.Now().UnixMilli(),
		Transactions: txs,
	}
	b.blockSvc.Queue(block)
}
