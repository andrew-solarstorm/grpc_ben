package main

import (
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/rs/zerolog/log"
)

// TransactionInfo is a struct that contains the transaction data
// Move to some where else?
type TransactionInfo struct {
	Signature   solana.Signature
	Transaction *solana.Transaction
	Meta        rpc.TransactionMeta
	Slot        uint64
}

func (u *TransactionInfo) Decode(upd *pb.SubscribeUpdateTransaction) {
	u.Signature = solana.SignatureFromBytes(upd.Transaction.Signature)
	u.Transaction = u.decodeTransaction(upd.Transaction.Transaction)
	u.Meta = u.decodeTransactionMeta(upd.Transaction.Meta)
	u.Slot = upd.Slot
}

func (u *TransactionInfo) decodeTransaction(tx *pb.Transaction) *solana.Transaction {
	sigs := make([]solana.Signature, len(tx.Signatures))
	for i, sig := range tx.Signatures {
		sigs[i] = solana.SignatureFromBytes(sig)
	}

	return &solana.Transaction{
		Signatures: sigs,
		Message:    u.decodeTransactionMessage(tx),
	}
}

func (u *TransactionInfo) decodeTransactionMessage(tx *pb.Transaction) solana.Message {
	accountKeys := make([]solana.PublicKey, len(tx.Message.AccountKeys))
	for i, a := range tx.Message.AccountKeys {
		accountKeys[i] = solana.PublicKeyFromBytes(a)
	}

	instructions := make([]solana.CompiledInstruction, len(tx.Message.Instructions))
	for i, ix := range tx.Message.Instructions {
		instructions[i] = u.decodeInstruction(ix)
	}

	return solana.Message{
		AccountKeys: accountKeys,
		Header: solana.MessageHeader{
			NumRequiredSignatures:       uint8(tx.Message.Header.NumRequiredSignatures),
			NumReadonlySignedAccounts:   uint8(tx.Message.Header.NumReadonlySignedAccounts),
			NumReadonlyUnsignedAccounts: uint8(tx.Message.Header.NumReadonlyUnsignedAccounts),
		},
		RecentBlockhash:     solana.HashFromBytes(tx.Message.RecentBlockhash),
		Instructions:        instructions,
		AddressTableLookups: u.decodeLookupTables(tx),
	}
}

func (u *TransactionInfo) decodeInstruction(ix *pb.CompiledInstruction) solana.CompiledInstruction {
	return solana.CompiledInstruction{
		ProgramIDIndex: uint16(ix.ProgramIdIndex),
		Accounts:       u.readUint8As16Slice(ix.Accounts),
		Data:           ix.Data,
	}
}

func (u *TransactionInfo) decodeLookupTables(tx *pb.Transaction) solana.MessageAddressTableLookupSlice {
	luts := make(solana.MessageAddressTableLookupSlice, len(tx.Message.AddressTableLookups))
	for i, lut := range tx.Message.AddressTableLookups {
		luts[i] = solana.MessageAddressTableLookup{
			AccountKey:      solana.PublicKeyFromBytes(lut.AccountKey),
			WritableIndexes: u.readUint8Slice(lut.WritableIndexes),
			ReadonlyIndexes: u.readUint8Slice(lut.ReadonlyIndexes),
		}
	}

	return luts
}

func (u *TransactionInfo) decodeTransactionMeta(tx *pb.TransactionStatusMeta) rpc.TransactionMeta {
	preTb := make([]rpc.TokenBalance, len(tx.PreTokenBalances))
	for i, b := range tx.PreTokenBalances {
		preTb[i] = u.decodeTokenBalance(b)
	}

	postTb := make([]rpc.TokenBalance, len(tx.PostTokenBalances))
	for i, b := range tx.PostTokenBalances {
		postTb[i] = u.decodeTokenBalance(b)
	}

	rewards := make([]rpc.BlockReward, len(tx.Rewards))
	for i, r := range tx.Rewards {
		rewards[i] = u.decodeReward(r)
	}

	innerInstructions := make([]rpc.InnerInstruction, len(tx.InnerInstructions))
	for i, ix := range tx.InnerInstructions {
		innerInstructions[i] = u.decodeInnerInstruction(ix)
	}

	return rpc.TransactionMeta{
		Err:               tx.Err.String(),
		Fee:               tx.Fee,
		PreBalances:       tx.PreBalances,
		PostBalances:      tx.PostBalances,
		InnerInstructions: innerInstructions,
		PreTokenBalances:  preTb,
		PostTokenBalances: postTb,
		LogMessages:       tx.LogMessages,
		Rewards:           rewards,
		LoadedAddresses:   rpc.LoadedAddresses{},
	}
}

func (u *TransactionInfo) decodeTokenBalance(b *pb.TokenBalance) rpc.TokenBalance {
	m, _ := solana.PublicKeyFromBase58(b.Mint)
	o, _ := solana.PublicKeyFromBase58(b.Owner)

	return rpc.TokenBalance{
		AccountIndex: uint16(b.AccountIndex),
		Owner:        &o,
		Mint:         m,
		UiTokenAmount: &rpc.UiTokenAmount{
			Amount:         b.UiTokenAmount.Amount,
			Decimals:       uint8(b.UiTokenAmount.Decimals),
			UiAmountString: b.UiTokenAmount.UiAmountString,
		},
	}
}

func (u *TransactionInfo) decodeInnerInstruction(b *pb.InnerInstructions) rpc.InnerInstruction {
	ixs := make([]rpc.CompiledInstruction, len(b.Instructions))
	for i, ix := range b.Instructions {
		ixs[i] = rpc.CompiledInstruction{
			ProgramIDIndex: uint16(ix.ProgramIdIndex),
			Accounts:       u.readUint8As16Slice(ix.Accounts),
			Data:           ix.Data,
		}
	}

	return rpc.InnerInstruction{
		Index:        uint16(b.Index),
		Instructions: ixs,
	}
}

func (u *TransactionInfo) decodeReward(b *pb.Reward) rpc.BlockReward {
	pk, _ := solana.PublicKeyFromBase58(b.Pubkey)

	return rpc.BlockReward{
		Pubkey:      pk,
		Lamports:    b.Lamports,
		PostBalance: b.PostBalance,
		RewardType:  rpc.RewardType(b.RewardType),
	}
}

func (u *TransactionInfo) readUint8As16Slice(data []byte) []uint16 {
	dec := bin.NewBinDecoder(data)
	var slice []uint16
	for {
		if !dec.HasRemaining() {
			break
		}
		ac, err := dec.ReadUint8()
		if err != nil {
			break
		}
		slice = append(slice, uint16(ac))
	}
	return slice
}

func (u *TransactionInfo) readUint8Slice(data []byte) []uint8 {
	dec := bin.NewBinDecoder(data)
	var slice []uint8
	for {
		if !dec.HasRemaining() {
			break
		}
		ac, err := dec.ReadUint8()
		if err != nil {
			break
		}
		slice = append(slice, ac)
	}
	return slice
}

type TxExtractArgs struct {
	AccountKeys   []solana.PublicKey
	Instructions  []rpc.CompiledInstruction
	Transaction   *TransactionInfo
	Prebalances   map[uint16]rpc.TokenBalance
	Postbalances  map[uint16]rpc.TokenBalance
	AccToMintMap  map[solana.PublicKey]solana.PublicKey
	AccToOwnerMap map[solana.PublicKey]solana.PublicKey

	TxContext *TxContext
}

func processTx(tx *TransactionInfo) *TxExtractArgs {
	accountKeys := tx.Transaction.Message.AccountKeys

	instructions := make([]rpc.CompiledInstruction, 0)

	for _, ix := range tx.Transaction.Message.Instructions {
		programID := accountKeys[ix.ProgramIDIndex]
		if programID.Equals(solana.TokenProgramID) {
			ixDataDecoded, err := base58.Decode(ix.Data.String())
			if err != nil {
				log.Error().Err(err).Msg("failed to decode instruction data from base58")
				continue
			}

			instructions = append(instructions, rpc.CompiledInstruction{
				ProgramIDIndex: ix.ProgramIDIndex,
				Accounts:       ix.Accounts,
				Data:           ixDataDecoded,
			})
		}
	}

	accToMintMap := make(map[solana.PublicKey]solana.PublicKey)
	accToOwnerMap := make(map[solana.PublicKey]solana.PublicKey)

	for _, inner := range tx.Meta.InnerInstructions {
		for _, ix := range inner.Instructions {
			if ix.ProgramIDIndex >= uint16(len(accountKeys)) {
				continue
			}
			programID := accountKeys[ix.ProgramIDIndex]
			if programID.Equals(solana.TokenProgramID) {
				instructions = append(instructions, ix)
			}
		}
	}

	preBalances := make(map[uint16]rpc.TokenBalance)
	postBalances := make(map[uint16]rpc.TokenBalance)

	for _, bal := range tx.Meta.PreTokenBalances {
		preBalances[bal.AccountIndex] = bal
		if int(bal.AccountIndex) >= len(accountKeys) {
			continue
		}
		key := accountKeys[bal.AccountIndex]
		accToMintMap[key] = bal.Mint
		if bal.Owner != nil {
			accToOwnerMap[key] = *bal.Owner
		}
	}
	for _, bal := range tx.Meta.PostTokenBalances {
		postBalances[bal.AccountIndex] = bal
		if int(bal.AccountIndex) >= len(accountKeys) {
			continue
		}
		key := accountKeys[bal.AccountIndex]
		accToMintMap[key] = bal.Mint
		if bal.Owner != nil {
			accToOwnerMap[key] = *bal.Owner
		}
	}

	// Create TxExtractArgs with all the extracted data
	args := &TxExtractArgs{
		AccountKeys:   accountKeys,
		Instructions:  instructions,
		Transaction:   tx,
		Prebalances:   preBalances,
		Postbalances:  postBalances,
		AccToMintMap:  accToMintMap,
		AccToOwnerMap: accToOwnerMap,
	}
	return args
}
