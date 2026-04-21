package main

import (
	"time"

	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

type TxContext struct {
	GeyserSentTime     time.Time
	ServerReceivedTime time.Time
	Slot               uint64
	BlockTime          int64

	BlockTx *pb.SubscribeUpdateTransactionInfo
}

// Transfer holds the structured data for a single SPL Token transfer.
type Transfer struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    uint64 `json:"amount"`
	Mint      string `json:"mint_address"`
	Signature string `json:"transaction_signature"`
	Slot      uint64 `json:"slot"`

	BlockTime int64 `json:"block_time"`

	GeyserSentTime     int64 `json:"geyser_sent_time"`
	ServerReceivedTime int64 `json:"server_received_time"`
	DecodedTime        int64 `json:"decoded_time"`
	WSSentTime         int64 `json:"ws_sent_time"`
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
