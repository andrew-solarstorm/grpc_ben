package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type Decoder struct {
	ch chan *TxContext

	wsSvc *WebsocketService
}

func NewTransferDecService(wsSvc *WebsocketService) *Decoder {
	dec := &Decoder{
		ch:    make(chan *TxContext, 100),
		wsSvc: wsSvc,
	}

	go dec.workers()
	return dec
}

func (svc *Decoder) Close() {
	close(svc.ch)
}

func (svc *Decoder) Queue(tx *TxContext) {
	svc.ch <- tx
}

func (svc *Decoder) workers() {
	for tx := range svc.ch {
		txInfo := &TransactionInfo{}
		txInfo.Decode(tx.Upd)

		decodeArgs := processTx(txInfo)
		decodeArgs.TxContext = tx

		for _, ix := range decodeArgs.Instructions {
			evt, _ := svc.Decode(decodeArgs, &ix)
			if evt != nil {
				svc.wsSvc.Send(evt.Mint, evt)
			}
		}

	}
}

func (svc *Decoder) Decode(args *TxExtractArgs, ixs *rpc.CompiledInstruction) (*Transfer, error) {
	if ixs == nil || len(ixs.Data) == 0 {
		return nil, nil
	}
	disc := ixs.Data[0]
	data := ixs.Data[1:]
	switch disc {
	case 3:
		return svc.decodeTransfer(args, ixs.Accounts, data)
	case 12:
		return svc.decodeTransferChecked(args, ixs.Accounts, data)
	}
	return nil, nil
}

func (svc *Decoder) decodeTransfer(args *TxExtractArgs, ixAccounts []uint16, ixData []byte) (*Transfer, error) {
	source, err := getFromAccountKeys(args.AccountKeys, ixAccounts, 0)
	if err != nil {
		return nil, err
	}

	destination, err := getFromAccountKeys(args.AccountKeys, ixAccounts, 1)
	if err != nil {
		return nil, err
	}

	if len(ixData) < 8 {
		return nil, errors.New("insufficient data for amount")
	}
	amount := binary.LittleEndian.Uint64(ixData[:8])

	mint := args.AccToMintMap[source]

	transfer := Transfer{
		From:               args.AccToOwnerMap[source].String(),
		To:                 args.AccToOwnerMap[destination].String(),
		Mint:               mint.String(),
		Amount:             amount,
		Signature:          args.Transaction.Signature.String(),
		Slot:               args.Transaction.Slot,
		BlockTime:          args.TxContext.BlockTime,
		GeyserSentTime:     args.TxContext.GeyserSentTime.UnixMicro(),
		ServerReceivedTime: args.TxContext.ServerReceivedTime.UnixMicro(),
		DecodedTime:        time.Now().UnixMicro(),
	}
	if bytes.Equal(mint[:], solana.SolMint[:]) {
		transfer.From = source.String()
		transfer.To = destination.String()
	}
	return &transfer, nil
}

func (svc *Decoder) decodeTransferChecked(args *TxExtractArgs, ixAccounts []uint16, ixData []byte) (*Transfer, error) {
	source, err := getFromAccountKeys(args.AccountKeys, ixAccounts, 0)
	if err != nil {
		return nil, err
	}

	mint, err := getFromAccountKeys(args.AccountKeys, ixAccounts, 1)
	if err != nil {
		return nil, err
	}

	destination, err := getFromAccountKeys(args.AccountKeys, ixAccounts, 2)
	if err != nil {
		return nil, err
	}

	if len(ixData) < 9 {
		return nil, errors.New("insufficient data for amount and decimals")
	}
	amount := binary.LittleEndian.Uint64(ixData[:8])

	transfer := Transfer{
		From:               args.AccToOwnerMap[source].String(),
		To:                 args.AccToOwnerMap[destination].String(),
		Mint:               mint.String(),
		Amount:             amount,
		Signature:          args.Transaction.Signature.String(),
		Slot:               args.Transaction.Slot,
		BlockTime:          args.TxContext.BlockTime,
		GeyserSentTime:     args.TxContext.GeyserSentTime.UnixMicro(),
		ServerReceivedTime: args.TxContext.ServerReceivedTime.UnixMicro(),
		DecodedTime:        time.Now().UnixMicro(),
	}
	if bytes.Equal(mint[:], solana.SolMint[:]) {
		transfer.From = source.String()
		transfer.To = destination.String()
	}

	return &transfer, nil
}

func getFromAccountKeys(accountKeys []solana.PublicKey, ixAccounts []uint16, index int) (solana.PublicKey, error) {
	if index >= len(ixAccounts) {
		return solana.PublicKey{}, errors.New("index out of range")
	}
	globalIndex := ixAccounts[index]
	if int(globalIndex) >= len(accountKeys) {
		return solana.PublicKey{}, errors.New("index out of range")
	}
	return accountKeys[globalIndex], nil
}
