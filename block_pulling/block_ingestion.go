package main

import (
	"context"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

var PROGRAM_IDS = []string{
	"opnb2LAfJYbRMAHHvqjCwQxanZn7ReEHp1k81EohpZb",
	"DEXYosS6oEGvk8uCDayvwEZz4qEyDJRf9nFgYCaqPMTm",
	"AMM55ShdkoGRB5jVYPjWziwk8m5MpwyDgsMWHaMSQWH6",
	"CURVGoZn8zycx6FXwwevgBTB2gVvdbGTEpvMJDbgs2t4",
	"D3BBjqUdCYuP18fNvvMbPAZ8DpcRi4io2EsYHQawJDag",
	"BSwp6bEBihVLdqJRKGgzjcGLHkcTuzmSo1TQkHepzH8p",
	"C1onEW2kPetmHmwe74YC1ESx3LnFEpVau6g2pg4fHycr",
	"6MLxLqiXaaSUpkgMnWDTuejNZEz3kE7k2woyHGVFw319",
	"CLMM9tUoggJu2wagPkkqs9eFG4BWhVBZWkP1qv3Sp7tR",
	"H8W3ctz92svYg6mkn1UtGfu2aQr2fnUFHM1RhScEtQDt",
	"CTMAxxk34HjKWxQ3QLZK1HpaLXmBveao3ESePXbiyfzh",
	"cysPXAjehMpVKUapzbMCCnpFxUFFryEWEaLgnb9NrR8",
	"GNExJhNUhc9LN2DauuQAUJnXoy6DJ6zey3t9kT9A2PF3",
	"DSwpgjMvXhtGn6BsbqmacdBZyfLj6jSWf3HJpdJtmg6N",
	"Dooar9JkhdZ7J3LHN3A7YCuoGRUggXhQaG4kijfLGU2j",
	"dp2waEWSBy5yKmq65ergoU3G6qRLmqa6K7We4rZSKph",
	"FLUXubRmkEi2q6K3Y9kBPg9248ggaZVsoSFhtJHSrm1X",
	"7WduLbRfYhTJktjLw5FDEyrqoEv61aTTCuGAetgLjzN5",
	"Gswppe6ERWKpUTXvRPfXdzHhiCyJvLadVvXGfdpBqcE1",
	"HyaB3W9q6XdA5xwpU4XnSZV94htfmbmqJXZcEbRaJutt",
	"PERPHjGBqRHArX4DySjwM6UJHiR3sWAatqfdBS2qQJu",
	"DCA265Vj8a9CEuX1eb1LWRnDT7uK6q1xMipnNyatn23M",
	"jupoNjAxXgZ4rjzxzPMP4oxduvQsQtZzyknqvzYNrNu",
	"CrX7kMhLC3cSsXJdT7JDgqrRVWGnUpX3gfEfxxU2NVLi",
	"EewxydAPCCVuNEyrVN68PuSYdQ7wKn27V9Gjeoi8dy3S",
	"2wT8Yq49kHgDzXuPxZSaeLaH1qbmGXtEyPy64bL7aD3c",
	"9tKE7Mbmj4mxDjWatikzGAtkoWosiiZX9y6J4Hfm2R8H",
	"MarBmsSgKXdrN1egZf5sqe1TMai9K1rChYNDJgjq7aD",
	"MERLuDFBMmsHnsBPZw2sDQZHvXFMwp8EdjudcU2HKky",
	"5B23C376Kwtd1vzb5LCJHiHLPnoWSnnx661hhGGDEv8y",
	"Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB",
	"LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo",
	"1MooN32fuBBgApc8ujknKJw5sef3BVwPGgz3pto1BAh",
	"srmqPvymJeFKQ4zGQed1GFppgkRHL9kaELCbyksJtPX",
	"DjVE6JNiYqPL2QXyCUUh8rNjHrbz9hXHNYt99MQ59qw1",
	"9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP",
	"PSwapMdSai8tjrEXcxFeQth87xC4rRsa4VA5mhGhXkP",
	"PhoeNiXZ8ByJGLkxNfZRnkUfjvmuYqLR89jjFHGqdXY",
	"6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P",
	"CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK",
	"CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C",
	"675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8",
	"5quBtoiQqxF9Jv6KYKctB59NT3gtJD2Y65kdnB1Uev3h",
	"SSwpkEEcbUqx4vtoEByFjSkhKdCT862DNVb52nZg1UZ",
	"SP12tWFxD9oJsVWNavTTBZvMbA6gkAmxtVgxdqvyvhY",
	"SPMBzsVUuoHA4Jm6KunbsotaahvVikZs1JyTW6iJvbn",
	"SSwapUtytfBdBn1b9NUGG6foMVPtcWgpRU32HToDUZr",
	"SCHAtsf8mbjyjiv4LkhLKutTf6JnZAbdJKFkXQNMFHZ",
	"9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin",
	"5ocnV1qiCgaQR8Jb8xWnVbApfaygJ8tNoZfgPwsgx9kx",
	"Hsn6R7N5avWAL4ScKHYgmwFyhnQ7ZEun94AmTiptPRdA",
	"swapNyd8XiQwJ6ianp9snpu4brUqFxadzvHebnAXjJZ",
	"swapFpHZwjELNnjvThjajtiVmkz3yPQEHjLtka2fwHW",
	"SPoo1Ku8WFXoNDMHPsrGSTSG1Y47rzgn41SLUNakuHy",
	"SSwpMgqNDsyV7mAgN9ady4bDVu5ySjmmXejXvy2vLt1",
	"SwaPpA9LAaLfeLi3a68M4DjnLqgtticKg6CnyNwgAC8",
	"SWiMDJYFUGj6cPrQ6QYYYWZtvXQdRChSVAygDZDsCHC",
	"SWimmSE5hgWsEruwPBLBVAFi3KyVfe8URU2pb4w7GZs",
	"2KehYt3KsEQR53jYcxjbQp2d2kCp4AkuQW68atufRwSr",
	"SwAPNuiTrUSw3p96z3dUBW7d51ge8UiRsnWAtRLnF8e",
	"whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc",
	"XuErbiqKKqpvN2X8qjkBNo2BwNvQp1WZKZTDgxKB95r",
	"MNFSTqtC93rEfYHB6hF82sKdZpUDFWkViLByLd1k1Ms",
	"61DFfeTKM7trxYcPQCM78bJ794ddZprZpAwAnLiwTpYH",
	"MoonCVVNZFSYkqNXP6bxHLPL6QQJiMagDL3qcqUQTrG",
	"NUMERUNsFCP3kuNmWZuXtm1AaQCPj9uw6Guv2Ekoi5P",
	"j1o2qRpjcyUwEvwtcfhEQefh773ZgjxcVRry7LDqg5X",
	"ZERor4xhbUycZ6gb9ntrhqscUcZmAbQDjEAtCf4hbZY",
	"5jnapfrAN47UYkLkEf7HnprPPBCQLvkYWGZDeKkaP5hv",
	"SNaPnpKUY656VPwbKmKT8FG4T85g4VWhRH1B4TQUfKs",
	"FunojPVY4nWD7sFCBvQh2sSaTYbq4sUociswuBQfvFks",
	"5U3EU2ubXtK84QcRjWVmYt9RaDyA8gKxdUrPFXmZyaki",
	"PytERJFhAKuNNuaiXkApLfWzwNwSNDACpigT3LwQfou",
	"pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA",
	"super4XGGb7KWorPuoSNVQDHAVQjWzTpqcoRS86d9Us",
	"cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG",
	"virEFLZsQm1iFAs8py1XnziJ67gTzW2bfCWhxNPfccD",
	"LanMV9sAd7wArD4vJFi2qDdfnVhFxYSUg6eADduJ3uj",
	"dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN",
	"vrTGoBuy5rYSxAfV3jaRJWHH6nN9WK4NRExGxsk1bCJ",
	"srAMMzfVHVAtgSJc8iH6CfKzuWuUTzLHVCE81QU1rgi",
	"waveQX2yP3H1pVU8djGvEHmYg8uamQ84AuyGtpsrXTF",
	"45iBNkaENereLKMjLm2LHkF3hpDapf6mnvrM5HWFg9cY",
	"REALQqNEomY6cQGZJUGwywTBD2UmDT32rZcNnfxQ5N2",
	"1qbkdrr3z4ryLA7pZykqxvxWPoeifcVKo6ZG9CfkvVE",
}

type BlockIngestionService struct {
	cli *yellowstone.GeyserGrpcClient

	dec *Decoder
}

func NewBlockIngestionService(dec *Decoder) *BlockIngestionService {
	return &BlockIngestionService{
		dec: dec,
	}
}

func (svc *BlockIngestionService) Close() {
	if svc.cli != nil {
		svc.cli.Close()
	}
}

func (svc *BlockIngestionService) Subscribe(endpoint, token string, commitment *pb.CommitmentLevel) {
	builder, err := yellowstone.BuildFromShared(endpoint)
	if err != nil {
		log.Fatalf("Error building client: %v", err)
	}

	clientBuilder := builder.XToken(token).KeepAliveWhileIdle(true).MaxDecodingMessageSize(100 * 1024 * 1024)
	grpcClient, err := clientBuilder.Connect(context.Background())
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	svc.cli = grpcClient

	includeTx := true
	includeAccounts := false
	includeEntries := false

	req := &pb.SubscribeRequest{
		Blocks: map[string]*pb.SubscribeRequestFilterBlocks{
			"block_filter": {
				IncludeTransactions: &includeTx,
				IncludeAccounts:     &includeAccounts,
				IncludeEntries:      &includeEntries,
				AccountInclude:      PROGRAM_IDS,
			},
		},
		Commitment: commitment,
	}
	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("Listening for block updates...")
	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Block:
			block := update.GetBlock()
			for _, tx := range block.GetTransactions() {
				if tx == nil {
					continue
				}

				arrived := time.Now()

				txCtx := TxContext{
					GeyserSentTime:     update.CreatedAt.AsTime(),
					ServerReceivedTime: arrived,
					Slot:               block.Slot,
					BlockTx:            tx,
					BlockTime:          block.BlockTime.Timestamp,
				}
				svc.dec.Queue(&txCtx)
			}
		default:
			return nil
		}
		return nil
	})
}
