package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"github.com/gin-gonic/gin"
)

func main() {
	var endpoint, token, commitmentStr, httpAddr string

	flag.StringVar(&endpoint, "endpoint", "http://localhost:10000", "grpc server url")
	flag.StringVar(&token, "token", "", "auth token")
	flag.StringVar(&commitmentStr, "commitment", "PROCESSED", "commitment level")
	flag.StringVar(&httpAddr, "http", "0.0.0.0:8080", "http listen address for websocket server")
	flag.Parse()

	if httpAddr != "" {
		// Allow common shorthands:
		// - "8080"  => "0.0.0.0:8080"
		// - ":8080" => "0.0.0.0:8080"
		if !strings.Contains(httpAddr, ":") {
			httpAddr = "0.0.0.0:" + httpAddr
		} else if strings.HasPrefix(httpAddr, ":") {
			httpAddr = "0.0.0.0" + httpAddr
		}
	}

	fmt.Printf("ENDPOINT: %s TOKEN: %s CommitmentLevel: %s HTTP: %s\n", endpoint, token, commitmentStr, httpAddr)

	if token == "" {
		fmt.Println("ERR: token is not set")
		return
	}

	commitment := getCommitmentLevel(commitmentStr)

	sysClock := NewSysClock()
	sysClock.subscribe(endpoint, token, commitment)

	wsSvc := NewWSService()
	dec := NewTransferDecService(wsSvc)

	blockBuilder := BlockFormer(dec)
	blockBuilder.ch = make(chan *LocalBlock, 100)
	go blockBuilder.worker()

	buff := &Buffer{
		mu:       &sync.RWMutex{},
		txs:      make([]*pb.SubscribeUpdateTransactionInfo, 0),
		clock:    sysClock,
		blockSvc: blockBuilder,
	}
	blockMeta := NewBlockMeta(buff)
	ingest := NewTxIngestService(buff)
	ingest.Subscribe(endpoint, token, commitment)
	blockMeta.Subscribe(endpoint, token, commitment)

	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/ws", wsSvc.Handler)
	router.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })

	httpSrv := &http.Server{
		Addr:              httpAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("http server error: %v\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ingest.Close()
	blockMeta.Close()
	httpSrv.Shutdown(shutdownCtx)
	fmt.Println("✅ block_pulling completed")
}

func getCommitmentLevel(commitment string) *pb.CommitmentLevel {
	result := pb.CommitmentLevel_PROCESSED
	switch commitment {
	case "FINALIZED", "finalized":
		result = pb.CommitmentLevel_FINALIZED
		return &result
	case "CONFIRMED", "confirmed":
		result = pb.CommitmentLevel_CONFIRMED
		return &result
	}
	return &result
}
