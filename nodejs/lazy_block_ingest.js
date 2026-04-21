const Client = require("@triton-one/yellowstone-grpc").default;
const { CommitmentLevel } = require("@triton-one/yellowstone-grpc");

const args = require("minimist")(process.argv.slice(2));

const GRPC_URL = args.url || "http://10.0.0.250:10000";
const GRPC_TOKEN = args.token || "";
const COMMITMENT = CommitmentLevel[args.commitment?.toUpperCase()] ?? CommitmentLevel.PROCESSED;
const WS_PORT = args.port || 8080;

const PROGRAM_IDS = [
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
];
// Map from slot (string or number) to tracking info
const blocksTracker = new Map();




async function subscribe() {
    const client = new Client(GRPC_URL, GRPC_TOKEN, {
        "grpc.max_receive_message_length": 64 * 1024 * 1024,
    });

    await client.connect();
    console.log("Connected to Yellowstone GRPC:", GRPC_URL);

    const stream = await client.subscribe();

    stream.on("data", (data) => {
        const nowUs = BigInt(Date.now()) * 1000n;

        // 1. Check for transaction
        if (data.transaction) {
            const slot = data.transaction.slot;
            if (slot) {
                if (!blocksTracker.has(slot)) {
                    blocksTracker.set(slot, {
                        firstTxTime: nowUs,
                        txCount: 1,
                    });
                } else {
                    const track = blocksTracker.get(slot);
                    if (!track.firstTxTime) {
                        track.firstTxTime = nowUs;
                    }
                    track.txCount++;
                }
            }
        }

        const m = data.blockmeta || data.blockMeta;
        // 2. Check for block_meta
        if (m) {
            const slot = m.slot;
            if (slot) {

                let track = blocksTracker.get(slot);
                if (!track) {
                    track = {
                        firstTxTime: 0,
                        blockMetaTime: nowUs,
                        txCount: 0,
                        blockTime: m.blockTime?.timestamp
                    };
                    blocksTracker.set(slot, track);
                } else {
                    track.blockTime = m.blockTime?.timestamp;
                    track.blockMetaTime = nowUs;
                }

                const delaysMs = track.firstTxTime ? Number(nowUs - track.firstTxTime) / 1000 : 0;
                const blockTimeStr = track.blockTime ? new Date(Number(track.blockTime) * 1000).toISOString() : 'N/A';
                console.log(`Slot: ${slot} | Txs so far: ${track.txCount} | Time from first tx to block_meta: ${delaysMs}ms | block_time: ${track.blockTime} (${blockTimeStr})`);

                // Clean up slot
                blocksTracker.delete(slot);

                // Also clean up old slots to prevent memory leak
                for (const [s, data] of blocksTracker.entries()) {
                    // If a slot has been tracked for > 30s we can probably delete it
                    if (data.firstTxTime && nowUs - data.firstTxTime > 30000000n) {
                        blocksTracker.delete(s);
                    }
                }
            }
        }
    });

    // Subscribe to transactions AND block meta
    const req = {
        accounts: {},
        slots: {},
        transactions: {
            txSub: {
                accountInclude: PROGRAM_IDS,
                accountExclude: [],
                accountRequired: []
            }
        },
        transactionsStatus: {},
        entry: {},
        blocks: {},
        blocksMeta: {
            metaSub: {}
        },
        accountsDataSlice: [],
        ping: undefined,
        commitment: COMMITMENT,
    };

    stream.write(req);
    console.log("Subscribed to transactions and blockMeta stream...");
}

subscribe().catch((err) => {
    console.error("Error subscribing:", err);
});
