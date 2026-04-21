const { randomUUID } = require('crypto');
const { WSMsgAction } = require('./types');
const { PublicKey } = require('@solana/web3.js');
const WebSocket = require('ws');

class Client {
    constructor(wsSvc, id, conn) {
        this.ID = id;
        this.conn = conn;
        this.mints = new Set();
        this.wsSvc = wsSvc;
        
        // Handle read operations similarly to readWorker explicitly passing to logic handler layer
        this.conn.on('message', (msg) => this.readWorker(msg));
        this.conn.on('close', () => {
            // Unexpected closures might be logged but here we just immediately drop all sub state mapping UnSubscribe
            for (const mint of this.mints) {
                this.wsSvc.UnSubscribe(mint, this);
            }
        });
        this.conn.on('error', (err) => {
            // websocket error handle
        });
    }

    Send(msg) {
        if (this.conn.readyState === WebSocket.OPEN || this.conn.readyState === 1) { 
            this.conn.send(msg);
        }
    }

    readWorker(msg) {
        try {
            const cliMsg = JSON.parse(msg.toString());
            
            try {
                new PublicKey(cliMsg.mint);
            } catch(e) {
                console.error("invalid pk");
                return;
            }

            this.handle(cliMsg.action, cliMsg.mint);
        } catch(err) {
            console.error('Failed to parse client message:', err);
        }
    }

    handle(action, mint) {
        switch (action) {
            case WSMsgAction.SUBSCRIBE:
                if (this.mints.has(mint)) {
                    return;
                }
                this.mints.add(mint);
                this.wsSvc.Subscribe(mint, this);
                break;
            case WSMsgAction.UNSUBSCRIBE:
                this.wsSvc.UnSubscribe(mint, this);
                this.mints.delete(mint);
                break;
        }
    }
}

class WebsocketService {
    constructor() {
        this.mintMap = new Map(); // string -> Map<uuid, Client>
        this.clientMap = new Map(); // uuid -> Client
    }

    handleConnection(conn) {
        const cid = randomUUID();
        console.log(`🔌 New WS client connection: ${cid}`);
        const client = new Client(this, cid, conn);
        this.clientMap.set(cid, client);
    }

    Subscribe(mint, client) {
        if (!this.mintMap.has(mint)) {
            this.mintMap.set(mint, new Map());
        }
        this.mintMap.get(mint).set(client.ID, client);
    }

    UnSubscribe(mint, client) {
        if (!this.mintMap.has(mint)) {
            return;
        }
        this.mintMap.get(mint).delete(client.ID);
        // Do not delete map entry on empty to prevent re-creation thrashing logic mapping exactly match Go maps which are retained globally
    }

    Send(mint, evt) {
        const clients = this.mintMap.get(mint);
        
        evt.ws_sent_time = Math.floor(Date.now() * 1000);

        if (clients) {
            const msgString = JSON.stringify(evt);
            clients.forEach(c => {
                c.Send(msgString);
            });
        }
    }
}

module.exports = {
    WebsocketService,
    Client
};
