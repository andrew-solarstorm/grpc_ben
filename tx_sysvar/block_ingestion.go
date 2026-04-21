package main

import (
	"time"

	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

type LocalBlock struct {
	CreatedAt    int64
	Blocktime    int64
	Slot         uint64
	Transactions []*pb.SubscribeUpdateTransactionInfo
}

type BlockBuilder struct {
	ch  chan *LocalBlock
	dec *Decoder
}

func BlockFormer(dec *Decoder) *BlockBuilder {
	svc := &BlockBuilder{
		dec: dec,
		ch:  make(chan *LocalBlock, 100),
	}
	go svc.worker()
	return svc
}

func (b *BlockBuilder) Queue(block *LocalBlock) {
	b.ch <- block
}
func (b *BlockBuilder) worker() {
	for block := range b.ch {
		now := time.Now()
		for _, tx := range block.Transactions {
			if tx == nil {
				continue
			}

			txCtx := TxContext{
				GeyserSentTime:     now,
				ServerReceivedTime: now,
				Slot:               block.Slot,
				BlockTx:            tx,
				BlockTime:          block.Blocktime,
			}
			b.dec.Queue(&txCtx)
		}
	}
}
