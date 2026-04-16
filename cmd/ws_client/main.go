package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
)

type WSMsgAction int8

const (
	WSMsgAction_SUBSCRIBE WSMsgAction = iota
	WSMsgAction_UNSUBSCRIBE
)

type ClientSubMsg struct {
	Action WSMsgAction `json:"action"`
	Mint   string      `json:"mint"`
}

// Matches server `Transfer` JSON payload (types.go)
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

func main() {
	var wsURL, mint string
	var showRaw bool

	flag.StringVar(&wsURL, "ws", "ws://localhost:8080/ws", "websocket url")
	flag.StringVar(&mint, "mint", "", "mint base58 to subscribe to")
	flag.BoolVar(&showRaw, "raw", false, "print raw json line too")
	flag.Parse()

	if mint == "" {
		fmt.Println("ERR: -mint is required")
		os.Exit(2)
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Printf("dial error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	sub := ClientSubMsg{Action: WSMsgAction_SUBSCRIBE, Mint: mint}
	subBytes, _ := sonic.Marshal(&sub)
	if err := conn.WriteMessage(websocket.TextMessage, subBytes); err != nil {
		fmt.Printf("subscribe write error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("subscribed: mint=%s ws=%s\n", mint, wsURL)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		unsub := ClientSubMsg{Action: WSMsgAction_UNSUBSCRIBE, Mint: mint}
		if b, e := sonic.Marshal(&unsub); e == nil {
			_ = conn.WriteMessage(websocket.TextMessage, b)
		}
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
		_ = conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("read error: %v\n", err)
			return
		}
		if showRaw {
			fmt.Printf("raw=%s\n", string(msg))
		}

		var t Transfer
		if err := sonic.Unmarshal(msg, &t); err != nil {
			fmt.Printf("unmarshal error: %v msg=%s\n", err, string(msg))
			continue
		}

		now := time.Now()
		nowUS := now.UnixMicro()
		nowS := now.Unix()
		// All times are unix microseconds
		recvDelayUS := t.ServerReceivedTime - t.GeyserSentTime
		decodeDelayUS := t.DecodedTime - t.ServerReceivedTime
		totalDelayUS := t.DecodedTime - t.GeyserSentTime
		decodeToWSUS := t.WSSentTime - t.DecodedTime
		wsToClientUS := nowUS - t.WSSentTime
		clientLagUS := nowUS - t.DecodedTime

		// BlockTime is unix seconds (from geyser block meta), may be 0 when unknown
		var blockLagUS int64 = -1
		var blockLagS float64
		var blockToClientUS int64 = -1
		var blockToClientS int64 = -1
		if t.BlockTime > 0 {
			blockLagUS = t.DecodedTime - (t.BlockTime * 1_000_000)
			blockLagS = float64(blockLagUS) / 1_000_000.0
			blockToClientUS = nowUS - (t.BlockTime * 1_000_000)
			blockToClientS = nowS - t.BlockTime
		}

		fmt.Printf(
			"sig=%s slot=%d mint=%s amount=%d from=%s to=%s\n",
			t.Signature, t.Slot, t.Mint, t.Amount, t.From, t.To,
		)
		fmt.Printf(
			"  times(us): geyser_sent=%d server_recv=%d decoded=%d ws_sent=%d client_now=%d block_time_s=%d\n",
			t.GeyserSentTime, t.ServerReceivedTime, t.DecodedTime, t.WSSentTime, nowUS, t.BlockTime,
		)
		fmt.Printf(
			"  delays: recv=%s decode=%s decode_to_ws=%s ws_to_client=%s total=%s client_lag=%s",
			time.Duration(recvDelayUS)*time.Microsecond,
			time.Duration(decodeDelayUS)*time.Microsecond,
			time.Duration(decodeToWSUS)*time.Microsecond,
			time.Duration(wsToClientUS)*time.Microsecond,
			time.Duration(totalDelayUS)*time.Microsecond,
			time.Duration(clientLagUS)*time.Microsecond,
		)
		fmt.Printf(" block_lag=%s (%.3fs)", time.Duration(blockLagUS)*time.Microsecond, blockLagS)
		fmt.Printf(" block_lag_s=%f", blockLagS)
		fmt.Printf(" block_to_client=%s", time.Duration(blockToClientUS)*time.Microsecond)
		fmt.Printf(" block_to_client_s=%d", blockToClientS)
		fmt.Printf("\n\n")
	}
}
