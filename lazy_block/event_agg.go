package main

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type EventAggregator struct {
	mu      sync.RWMutex
	pending map[uint64][]*Transfer

	wsSvc *WebsocketService
}

func NewEventAggregator(wsSvc *WebsocketService) *EventAggregator {
	return &EventAggregator{
		pending: make(map[uint64][]*Transfer),
		wsSvc:   wsSvc,
	}
}

func (svc *EventAggregator) Push(evt *Transfer) {
	if evt == nil {
		return
	}
	if evt.BlockTime == 0 {
		svc.mu.Lock()
		_, ok := svc.pending[evt.Slot]
		if !ok {
			svc.pending[evt.Slot] = make([]*Transfer, 0)
		}
		svc.pending[evt.Slot] = append(svc.pending[evt.Slot], evt)
		svc.mu.Unlock()

		return
	}

	if evt.BlockTime > 0 && len(svc.pending[evt.Slot]) > 0 {
		svc.mu.Lock()
		log.Info().Uint64("slot", evt.Slot).Int64("block_time", evt.BlockTime).Int("pending", len(svc.pending[evt.Slot])).Msg("Push pending transfers")

		pending := svc.pending[evt.Slot]
		delete(svc.pending, evt.Slot)
		svc.mu.Unlock()

		for _, xfer := range pending {
			xfer.BlockTime = evt.BlockTime
			svc.wsSvc.Send(xfer.Mint, xfer)
		}
	}

	svc.wsSvc.Send(evt.Mint, evt)
}

func (svc *EventAggregator) OnBlockMeta(slot uint64, blockTime int64) {
	log.Info().Uint64("slot", slot).Int64("block_time", blockTime).Msg("OnBlockMeta")
	if blockTime <= 0 {
		return
	}

	svc.mu.Lock()
	pending := svc.pending[slot]
	if len(pending) > 0 {
		delete(svc.pending, slot)
	}
	svc.mu.Unlock()

	for _, evt := range pending {
		evt.BlockTime = blockTime
		svc.wsSvc.Send(evt.Mint, evt)
	}
}
