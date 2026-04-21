const yellowstone = require('@triton-one/yellowstone-grpc');

function decodeClock(data) {
    if (!data || data.length < 40) return null;
    const view = new DataView(
        data.buffer || data,
        data.byteOffset || 0,
        data.byteLength || data.length
    );
    const slot = Number(view.getBigUint64(0, true));
    const epochStartTimestamp = Number(view.getBigInt64(8, true));
    const epoch = Number(view.getBigUint64(16, true));
    const leaderScheduleEpoch = Number(view.getBigUint64(24, true));
    const unixTimestamp = Number(view.getBigInt64(32, true));
    return { slot, epochStartTimestamp, epoch, leaderScheduleEpoch, unixTimestamp };
}

class LRUCache {
    constructor(capacity) {
        this.capacity = capacity;
        this.cache = new Map();
    }
    get(key) {
        if (!this.cache.has(key)) return 0; // Go lru returns 0 value default for absent ints
        const val = this.cache.get(key);
        this.cache.delete(key);
        this.cache.set(key, val);
        return val;
    }
    has(key) {
        return this.cache.has(key);
    }
    add(key, value) {
        if (this.cache.has(key)) {
            this.cache.delete(key);
        } else if (this.cache.size >= this.capacity) {
            const firstKey = this.cache.keys().next().value;
            this.cache.delete(firstKey);
        }
        this.cache.set(key, value);
    }
}

class SystemClock {
    constructor() {
        this.cli = null;
        this.slots = new LRUCache(30);
    }

    TimeStamp(slot) {
        return this.slots.get(slot) || 0;
    }

    Close() {
        if (this.cli) {
            this.cli.close && this.cli.close();
            this.cli = null;
        }
    }

    async subscribe(endpoint, token, commitment) {
        let commitmentLevel;
        if (commitment === 'FINALIZED') {
            commitmentLevel = yellowstone.CommitmentLevel.FINALIZED;
        } else if (commitment === 'CONFIRMED') {
            commitmentLevel = yellowstone.CommitmentLevel.CONFIRMED;
        } else {
            commitmentLevel = yellowstone.CommitmentLevel.PROCESSED;
        }

        try {
            const client = new yellowstone.default(endpoint, token, {
                "grpc.max_receive_message_length": 64 * 1024 * 1024,
                "grpc.keepalive_time_ms": 10000,
            });
            this.cli = client;
            await this.cli.connect();

            const stream = await client.subscribe();

            const req = {
                slots: {},
                accounts: {
                    sysvar_filter: {
                        account: ["SysvarC1ock11111111111111111111111111111111"],
                        owner: [],
                        filters: []
                    }
                },
                transactions: {},
                transactionsStatus: {},
                entry: {},
                blocksMeta: {},
                accountsDataSlice: [],
                ping: undefined,
                blocks: {},
                commitment: commitmentLevel
            };

            await new Promise((resolve, reject) => {
                stream.write(req, (err) => {
                    if (err) {
                        reject(err);
                    } else {
                        resolve();
                    }
                });
            });

            stream.on('data', (update) => {
                if (!update || !update.account || !update.account.account) return;
                
                const clock = decodeClock(update.account.account.data);
                if (clock) {
                    if (!this.slots.has(clock.slot)) {
                        this.slots.add(clock.slot, clock.unixTimestamp);
                    }
                }
            });

            stream.on('error', (err) => {
                console.error("SystemClock stream error:", err);
            });

        } catch (err) {
            console.error("Error subscribing system clock:", err);
            process.exit(1);
        }
    }
}

module.exports = { SystemClock };
