const { TxContext } = require('./types');

class LocalBlock {
    constructor(slot, blockTime, createdAt, transactions) {
        this.CreatedAt = createdAt;
        this.Blocktime = blockTime;
        this.Slot = slot;
        this.Transactions = transactions;
    }
}

class BlockBuilder {
    constructor(dec) {
        this.dec = dec;
        this.ch = [];
        this.isWorking = false;
    }

    Queue(block) {
        this.ch.push(block);
        this.worker(); 
    }

    async worker() {
        if (this.isWorking) return;
        this.isWorking = true;

        while (this.ch.length > 0) {
            const block = this.ch.shift();
            const now = new Date();

            for (const tx of block.Transactions) {
                if (!tx) continue;

                const txCtx = new TxContext();
                txCtx.GeyserSentTime = now;
                txCtx.ServerReceivedTime = now;
                txCtx.Slot = block.Slot;
                txCtx.BlockTx = tx;
                txCtx.BlockTime = block.Blocktime;

                this.dec.Queue(txCtx);
            }
            // Yield context so javascript runtime resembles go routines
            await new Promise(resolve => setImmediate(resolve));
        }

        this.isWorking = false;
    }
}

module.exports = {
    LocalBlock,
    BlockBuilder
};
