const bs58Raw = require('bs58');
const bs58 = bs58Raw.default || bs58Raw;

class Decoder {
    constructor(wsSvc) {
        this.wsSvc = wsSvc;
        // Token program ID
        this.TOKEN_PROGRAM_ID = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA";
    }

    decodeBlock(slot, transactions, firstTxTime, blockMetaTime, blockTime = 0) {
        for (const txData of transactions) {
            try {
                this.processTx(slot, txData, firstTxTime, blockMetaTime, blockTime);
            } catch (err) {
                console.error(`Error decoding tx for slot ${slot}:`, err.message);
            }
        }
    }

    processTx(slot, txData, firstTxTime, blockMetaTime, blockTime = 0) {
        const txInfo = txData;
        if (!txInfo) {
            console.log("No txInfo");
            return;
        }

        const meta = txInfo.meta;
        const msg = txInfo.transaction?.message;
        if (!meta || !msg) {
            console.log("Missing meta or msg", !!meta, !!msg, Object.keys(txInfo));
            return;
        }

        const signature = bs58.encode(txInfo.signature);
        const accountKeys = msg.accountKeys.map(k => bs58.encode(k));

        const accToMintMap = new Map();
        const accToOwnerMap = new Map();

        const processBalances = (balances) => {
            if (!balances) return;
            for (const bal of balances) {
                if (bal.accountIndex >= accountKeys.length) continue;
                const key = accountKeys[bal.accountIndex];
                accToMintMap.set(key, bal.mint);
                if (bal.owner) {
                    accToOwnerMap.set(key, bal.owner);
                }
            }
        };

        processBalances(meta.preTokenBalances);
        processBalances(meta.postTokenBalances);

        const instructions = [];
        if (msg.instructions) {
            instructions.push(...msg.instructions);
        }
        if (meta.innerInstructions) {
            for (const inner of meta.innerInstructions) {
                if (inner.instructions) {
                    instructions.push(...inner.instructions);
                }
            }
        }

        for (const ix of instructions) {
            const programID = accountKeys[ix.programIdIndex];
            if (programID !== this.TOKEN_PROGRAM_ID) continue;

            const ixData = Buffer.from(ix.data);
            if (ixData.length === 0) continue;

            const disc = ixData[0];
            let transferEvent = null;

            if (disc === 3) {
                transferEvent = this.decodeTransfer(accountKeys, accToMintMap, accToOwnerMap, ix, ixData, signature, slot, firstTxTime, blockMetaTime, blockTime);
            } else if (disc === 12) {
                transferEvent = this.decodeTransferChecked(accountKeys, accToMintMap, accToOwnerMap, ix, ixData, signature, slot, firstTxTime, blockMetaTime, blockTime);
            }

            if (transferEvent) {
                this.wsSvc.send(transferEvent.mint_address, transferEvent);
            }
        }
    }

    decodeTransfer(accountKeys, accToMintMap, accToOwnerMap, ix, ixData, signature, slot, firstTxTime, blockMetaTime, blockTime = 0) {
        if (ixData.length < 9) return null; // 1 byte disc + 8 bytes amount
        const sourceIdx = ix.accounts[0];
        const destIdx = ix.accounts[1];

        if (sourceIdx >= accountKeys.length || destIdx >= accountKeys.length) return null;

        const source = accountKeys[sourceIdx];
        const dest = accountKeys[destIdx];

        const amount = ixData.readBigUInt64LE(1).toString();
        const mint = accToMintMap.get(source);

        if (!mint) return null;

        let from = accToOwnerMap.get(source) || source;
        let to = accToOwnerMap.get(dest) || dest;

        return {
            from,
            to,
            amount: parseInt(amount, 10),
            mint_address: mint,
            transaction_signature: signature,
            slot: parseInt(slot, 10),
            block_time: blockTime,
            first_tx_time: Number(firstTxTime || 0),
            block_meta_time: Number(blockMetaTime || 0),
            ws_sent_time: Date.now() * 1000,
        };
    }

    decodeTransferChecked(accountKeys, accToMintMap, accToOwnerMap, ix, ixData, signature, slot, firstTxTime, blockMetaTime, blockTime = 0) {
        if (ixData.length < 9) return null; // 1 byte disc + 8 bytes amount (+ decimals)
        const sourceIdx = ix.accounts[0];
        const mintIdx = ix.accounts[1];
        const destIdx = ix.accounts[2];

        if (sourceIdx >= accountKeys.length || mintIdx >= accountKeys.length || destIdx >= accountKeys.length) return null;

        const source = accountKeys[sourceIdx];
        const mint = accountKeys[mintIdx];
        const dest = accountKeys[destIdx];

        const amount = ixData.readBigUInt64LE(1).toString();

        let from = accToOwnerMap.get(source) || source;
        let to = accToOwnerMap.get(dest) || dest;

        return {
            from,
            to,
            amount: parseInt(amount, 10),
            mint_address: mint,
            transaction_signature: signature,
            slot: parseInt(slot, 10),
            block_time: blockTime,
            first_tx_time: Number(firstTxTime || 0),
            block_meta_time: Number(blockMetaTime || 0),
            ws_sent_time: Date.now() * 1000,
        };
    }
}

module.exports = Decoder;
