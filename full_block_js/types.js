const { PublicKey } = require('@solana/web3.js');

/**
 * TxContext holds context about the transaction and block
 */
class TxContext {
    constructor() {
        this.GeyserSentTime = null; // Date
        this.ServerReceivedTime = null; // Date
        this.Slot = 0;
        this.BlockTime = 0;
        this.BlockTx = null; // The gRPC transaction object
    }
}

/**
 * Transfer holds the structured data for a single SPL Token transfer.
 */
class Transfer {
    constructor() {
        this.from = "";
        this.to = "";
        this.amount = 0n; // BigInt or String depending on usage, keeping Number/BigInt convention
        this.mint_address = "";
        this.transaction_signature = "";
        this.slot = 0;
        this.block_time = 0;
        
        this.geyser_sent_time = 0;
        this.server_received_time = 0;
        this.decoded_time = 0;
        this.ws_sent_time = 0;
    }
}

const WSMsgAction = {
    SUBSCRIBE: 0,
    UNSUBSCRIBE: 1
};

class ClientSubMsg {
    constructor() {
        this.action = WSMsgAction.SUBSCRIBE;
        this.mint = "";
    }
}

module.exports = {
    TxContext,
    Transfer,
    WSMsgAction,
    ClientSubMsg
};
