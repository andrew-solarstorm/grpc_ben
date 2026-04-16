package main

import (
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{}

type Client struct {
	ID   uuid.UUID
	conn *websocket.Conn

	msgCh chan []byte

	mintMu sync.RWMutex
	mints  map[string]struct{}

	wsSvc *WebsocketService
}

func NewClient(wsSvc *WebsocketService, id uuid.UUID, conn *websocket.Conn) *Client {
	c := &Client{
		conn:  conn,
		ID:    id,
		msgCh: make(chan []byte, 256),
		mints: make(map[string]struct{}),
		wsSvc: wsSvc,
	}

	go c.readWorker()
	go c.writeWorker()

	return c
}

func (c *Client) Send(msg []byte) {
	c.msgCh <- msg
}

func (c *Client) writeWorker() {
	for msg := range c.msgCh {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Error().Err(err).Msg("Client::worker failed to send message to client")
		}
	}
}

func (c *Client) readWorker() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Debug().
					Err(err).
					Str("client_id", c.ID.String()).
					Msg("WebSocket connection closed unexpectedly")
			}
			break
		}

		var cliMsg ClientSubMsg
		if err := sonic.Unmarshal(msg, &cliMsg); err != nil {
			log.Error().
				Err(err).
				Str("client_id", c.ID.String()).
				Str("message", string(msg)).
				Msg("Failed to unmarshal client message")
			return
		}
		_, err = solana.PublicKeyFromBase58(cliMsg.Mint)
		if err != nil {
			log.Error().Err(err).Msg("invalid pk")
			return
		}
		c.handle(cliMsg.Action, cliMsg.Mint)

	}
}

func (c *Client) handle(action WSMsgAction, mint string) {
	switch action {
	case WSMsgAction_SUBSCRIBE:
		c.mintMu.Lock()
		_, ok := c.mints[mint]
		if ok {
			c.mintMu.Unlock()
			return
		}
		c.mints[mint] = struct{}{}
		c.mintMu.Unlock()

		c.wsSvc.Subscribe(mint, c)
	case WSMsgAction_UNSUBSCRIBE:
		c.wsSvc.UnSubscribe(mint, c)
		c.mintMu.Lock()
		delete(c.mints, mint)
		c.mintMu.Unlock()
	}
}

type WebsocketService struct {
	mu      sync.RWMutex
	mintMap map[string]map[uuid.UUID]*Client

	cliMu     sync.RWMutex
	clientMap map[uuid.UUID]*Client
}

func NewWSService() *WebsocketService {
	return &WebsocketService{
		mintMap:   make(map[string]map[uuid.UUID]*Client),
		clientMap: make(map[uuid.UUID]*Client),
	}
}

func (svc *WebsocketService) Handler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, c.Request.Header)
	if err != nil {
		return
	}

	svc.cliMu.Lock()
	defer svc.cliMu.Unlock()

	cid := uuid.New()
	svc.clientMap[cid] = NewClient(svc, cid, conn)
}

func (svc *WebsocketService) Subscribe(mint string, client *Client) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	_, ok := svc.mintMap[mint]
	if !ok {
		svc.mintMap[mint] = make(map[uuid.UUID]*Client)
	}
	svc.mintMap[mint][client.ID] = client
}

func (svc *WebsocketService) UnSubscribe(mint string, cli *Client) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	_, ok := svc.mintMap[mint]
	if !ok {
		return
	}
	delete(svc.mintMap[mint], cli.ID)
}

func (svc *WebsocketService) Send(mint string, evt *Transfer) {
	svc.cliMu.RLock()
	defer svc.cliMu.RUnlock()

	clients, ok := svc.mintMap[mint]
	nowUS := time.Now().UnixMicro()
	payload := *evt
	payload.WSSentTime = nowUS

	bytes, err := sonic.Marshal(&payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal data")
		return
	}
	if ok {
		for _, c := range clients {
			c.Send(bytes)
		}
	}
}
