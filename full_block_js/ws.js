const { randomUUID } = require('crypto');
const { WSMsgAction } = require('./types');
const { PublicKey } = require('@solana/web3.js');

class Client {
    constructor(wsSvc, id, conn) {
        this.ID = id;
        this.conn = conn;
        this.mints = new Set();
        this.wsSvc = wsSvc;
        
        // In Node.js ws library, we just use event listeners instead of gochannels/workers
        this.conn.on('message', (msg) => this.readWorker(msg));
        this.conn.on('close', () => {
            for (const mint of this.mints) {
                this.wsSvc.UnSubscribe(mint, this);
            }
        });
        this.conn.on('error', (err) => {
            console.error(`WebSocket error for client ${this.ID}:`, err);
        });
    }

    Send(msg) {
        if (this.conn.readyState === 1) { // WebSocket.OPEN
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
        this.mintMap = new Map(); // Map<string, Map<string, Client>>
        this.clientMap = new Map(); // Map<string, Client>
    }

    // Handles upgrade event or connection directly from express-ws or standard ws server
    handleConnection(conn) {
        const cid = randomUUID();
        const client = new Client(this, cid, conn);
        this.clientMap.set(cid, client);
        return client;
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
        if (this.mintMap.get(mint).size === 0) {
            this.mintMap.delete(mint);
        }
    }

    Send(mint, evt) {
        const clients = this.mintMap.get(mint);
        
        evt.ws_sent_time = Date.now() * 1000; // Unix Microseconds equivalent loosely

        const msgString = JSON.stringify(evt);
        if (clients) {
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
