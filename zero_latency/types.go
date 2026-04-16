package main

import (
	"time"

	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

type TxContext struct {
	// Timing Fields
	GeyserSentTime     time.Time
	ServerReceivedTime time.Time
	Slot               uint64
	BlockTime          int64

	Upd *pb.SubscribeUpdateTransaction
}

// Transfer holds the structured data for a single SPL Token transfer.
type Transfer struct {
	// these will be filled at the time tx arrived
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    uint64 `json:"amount"`
	Mint      string `json:"mint_address"`
	Signature string `json:"transaction_signature"`
	Slot      uint64 `json:"slot"`

	// lazy update for better delay
	BlockTime int64 `json:"block_time"` // From the transaction data itself

	// Timing Fields
	GeyserSentTime     int64 `json:"geyser_sent_time"`     // From `update.CreatedAt`
	ServerReceivedTime int64 `json:"server_received_time"` // From `time.Now()` on receipt
	DecodedTime        int64 `json:"decoded_time"`         // A new `time.Now()` captured right after decoding is complete
	WSSentTime         int64 `json:"ws_sent_time"`         // time.Now() right before WS broadcast
}

type WSMsgAction int8

const (
	WSMsgAction_SUBSCRIBE WSMsgAction = iota
	WSMsgAction_UNSUBSCRIBE
)

type ClientSubMsg struct {
	Action WSMsgAction `json:"action"`
	Mint   string      `json:"mint"`
}
