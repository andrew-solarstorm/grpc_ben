const { LocalBlock } = require('./block_ingestion');

const BatchSize = 10;

class Buffer {
    constructor(clock, blockSvc, batchSize = 10) {
        this.txs = [];
        this.slot = 0;
        this.clock = clock;
        this.blockSvc = blockSvc;
        this.batchSize = batchSize
        this.firstTxTime = performance.now()
    }

    Add(slot, tx) {
        // console.log("Adding tx to buffer", slot);

        if (slot < this.slot) {
            return;
        }

        if (slot > this.slot || this.txs.length >= this.batchSize) {
            const temp = this.txs;
            const currentSlot = this.slot;

            // Go spawns a goroutine: go b.buildBlock(b.slot, temp)
            setImmediate(() => {
                this.buildBlock(currentSlot, temp);
            });

            this.txs = [];
            if (slot > this.slot) {
                const duration = (performance.now() - this.firstTxTime).toFixed(3);
                process.stdout.write(`Time from first tx -> block formed: ${duration}ms\n`);
                this.firstTxTime = performance.now()
                this.slot = slot;
            }
        }
        this.txs.push(tx);
    }

    buildBlock(slot, txs) {
        console.log(`Building Batch: ${slot} | Len: ${txs.length}`);
        const block = new LocalBlock(
            slot,
            this.clock.TimeStamp(slot),
            Date.now(),
            txs
        );
        this.blockSvc.Queue(block);
    }
}

module.exports = { Buffer };
