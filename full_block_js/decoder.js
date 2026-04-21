const { Transfer } = require('./types');
const { TransactionInfo, processTx } = require('./txn_types');
const { PublicKey } = require('@solana/web3.js');

const SOL_MINT_STR = "So11111111111111111111111111111111111111112";
const SOL_MINT = new PublicKey(SOL_MINT_STR);

class Decoder {
    constructor(wsSvc) {
        this.wsSvc = wsSvc;
    }

    Queue(tx) {
        this.processWorker(tx);
    }

    async processWorker(tx) {
        if (!tx || !tx.BlockTx) {
            return;
        }

        const txInfo = new TransactionInfo();
        txInfo.decodeFromBlock(tx.BlockTx, tx.Slot);

        const decodeArgs = processTx(txInfo);
        decodeArgs.TxContext = tx;

        for (const ix of decodeArgs.Instructions) {
            const evt = this.Decode(decodeArgs, ix);
            if (evt) {
                this.wsSvc.Send(evt.mint_address, evt);
            }
        }
    }

    Decode(args, ixs) {
        if (!ixs || !ixs.Data || ixs.Data.length === 0) {
            return null;
        }

        const disc = ixs.Data[0];
        const data = ixs.Data.slice(1);

        switch (disc) {
            case 3:
                return this.decodeTransfer(args, ixs.Accounts, data);
            case 12:
                return this.decodeTransferChecked(args, ixs.Accounts, data);
        }
        return null;
    }

    decodeTransfer(args, ixAccounts, ixData) {
        const source = this.getFromAccountKeys(args.AccountKeys, ixAccounts, 0);
        if (!source) return null;

        const destination = this.getFromAccountKeys(args.AccountKeys, ixAccounts, 1);
        if (!destination) return null;

        if (ixData.length < 8) {
            // insufficient data for amount
            return null;
        }

        // read Uint64 LE
        const amount = Number(this.readBigUInt64LE(ixData, 0));

        const sourceBase58 = source.toBase58();
        const destinationBase58 = destination.toBase58();
        const mint = args.AccToMintMap[sourceBase58];
        const ownerSource = args.AccToOwnerMap[sourceBase58];
        const ownerDest = args.AccToOwnerMap[destinationBase58];

        if (!mint) return null; // Wait... if there's no mint in pre/post balances, we might not get it here

        const transfer = new Transfer();
        transfer.from = ownerSource || sourceBase58; // In Go we stringified the owner pk or if missing, might error/empty
        transfer.to = ownerDest || destinationBase58;
        transfer.mint_address = typeof mint === 'string' ? mint : mint.toBase58();
        transfer.amount = amount;
        transfer.transaction_signature = args.Transaction.Transaction?.Signature || "";
        transfer.slot = args.Transaction.Transaction?.Slot || 0;
        transfer.block_time = args.TxContext.BlockTime;
        transfer.geyser_sent_time = Math.floor(args.TxContext.GeyserSentTime.getTime() * 1000);
        transfer.server_received_time = Math.floor(args.TxContext.ServerReceivedTime.getTime() * 1000);
        transfer.decoded_time = Date.now() * 1000;

        if (transfer.mint_address === SOL_MINT_STR) {
            transfer.from = sourceBase58;
            transfer.to = destinationBase58;
        }
        return transfer;
    }

    decodeTransferChecked(args, ixAccounts, ixData) {
        const source = this.getFromAccountKeys(args.AccountKeys, ixAccounts, 0);
        if (!source) return null;

        const mint = this.getFromAccountKeys(args.AccountKeys, ixAccounts, 1);
        if (!mint) return null;

        const destination = this.getFromAccountKeys(args.AccountKeys, ixAccounts, 2);
        if (!destination) return null;

        if (ixData.length < 9) { // 8 byte amount + 1 byte decimals
            return null;
        }

        const amount = Number(this.readBigUInt64LE(ixData, 0));
        const sourceBase58 = source.toBase58();
        const destinationBase58 = destination.toBase58();
        const ownerSource = args.AccToOwnerMap[sourceBase58];
        const ownerDest = args.AccToOwnerMap[destinationBase58];

        const transfer = new Transfer();
        transfer.from = ownerSource || sourceBase58;
        transfer.to = ownerDest || destinationBase58;
        transfer.mint_address = mint.toBase58();
        transfer.amount = amount;
        transfer.transaction_signature = args.Transaction.Transaction?.Signature || "";
        transfer.slot = args.Transaction.Transaction?.Slot || 0;
        transfer.block_time = args.TxContext.BlockTime;
        transfer.geyser_sent_time = Math.floor(args.TxContext.GeyserSentTime.getTime() * 1000);
        transfer.server_received_time = Math.floor(args.TxContext.ServerReceivedTime.getTime() * 1000);
        transfer.decoded_time = Date.now() * 1000;

        if (transfer.mint_address === SOL_MINT_STR) {
            transfer.from = sourceBase58;
            transfer.to = destinationBase58;
        }
        return transfer;
    }

    getFromAccountKeys(accountKeys, ixAccounts, index) {
        if (index >= ixAccounts.length) return null;
        const globalIndex = ixAccounts[index];
        if (globalIndex >= accountKeys.length) return null;
        return accountKeys[globalIndex];
    }

    readBigUInt64LE(buffer, offset) {
        // Read 8 bytes
        const first = buffer[offset];
        const last = buffer[offset + 7];
        if (first === undefined || last === undefined) {
            return 0n;
        }
        // Node.js Buffer has readBigUInt64LE, but Uint8Array does not natively have it, so we do it manually safely
        let val = 0n;
        for (let i = 0; i < 8; i++) {
            val += BigInt(buffer[offset + i]) << BigInt(8 * i);
        }
        return val;
    }
}

module.exports = {
    Decoder
};
