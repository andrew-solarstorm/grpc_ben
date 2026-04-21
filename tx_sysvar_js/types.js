class TxContext {
    constructor() {
        this.GeyserSentTime = null; // Date
        this.ServerReceivedTime = null; // Date
        this.Slot = 0; // Number
        this.BlockTime = 0; // Number

        this.BlockTx = null; // pb.SubscribeUpdateTransactionInfo
    }
}

class Transfer {
    constructor() {
        this.from = "";
        this.to = "";
        this.amount = 0;
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
        this.action = 0;
        this.mint = "";
    }
}

module.exports = {
    TxContext,
    Transfer,
    WSMsgAction,
    ClientSubMsg
};
