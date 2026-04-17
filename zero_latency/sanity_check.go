package main

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rs/zerolog/log"
)

type SanityChecker struct {
	mu    *sync.Mutex
	slots *lru.Cache[uint64, int64]
}

func NewSanityChecker() *SanityChecker {
	cache, _ := lru.New[uint64, int64](30)
	return &SanityChecker{
		mu:    new(sync.Mutex),
		slots: cache,
	}
}

func (c *SanityChecker) Check(slot uint64, timestamp int64, source string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if slotTime, ok := c.slots.Get(slot); ok {
		if slotTime != timestamp {
			log.Info().
				Str("Source", source).
				Uint64("Slot", slot).
				Int64("clocktime", slotTime).
				Int64("timestamp", timestamp).
				Msg("timestamp is not match")
		}
		return
	}

	c.slots.Add(slot, timestamp)
}
