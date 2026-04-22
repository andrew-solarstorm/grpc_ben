const WebSocket = require('ws');

class WebsocketService {
    constructor(port = 8080) {
        this.wss = new WebSocket.Server({ port });
        
        // mint -> Set of clients
        this.mintMap = new Map();

        this.wss.on('connection', (ws) => {
            ws.on('message', (message) => {
                try {
                    const data = JSON.parse(message);
                    if (!data.mint) return;
                    
                    if (data.action === 0) { // subscribe
                        this.subscribe(data.mint, ws);
                    } else if (data.action === 1) { // unsubscribe
                        this.unsubscribe(data.mint, ws);
                    }
                } catch (err) {
                    console.error("Failed to parse message", err);
                }
            });

            ws.on('close', () => {
                // When a connection closes, remove it from all subscriptions
                for (const [mint, clients] of this.mintMap.entries()) {
                    clients.delete(ws);
                    if (clients.size === 0) {
                        this.mintMap.delete(mint);
                    }
                }
            });
        });

        console.log(`WebSocket Server listening on port ${port}`);
    }

    subscribe(mint, ws) {
        if (!this.mintMap.has(mint)) {
            this.mintMap.set(mint, new Set());
        }
        this.mintMap.get(mint).add(ws);
        console.log(`Client subscribed to mint: ${mint}`);
    }

    unsubscribe(mint, ws) {
        if (this.mintMap.has(mint)) {
            this.mintMap.get(mint).delete(ws);
            if (this.mintMap.get(mint).size === 0) {
                this.mintMap.delete(mint);
            }
        }
    }

    send(mint, evt) {
        if (this.mintMap.has(mint)) {
            const clients = this.mintMap.get(mint);
            
            // Populate ws sent time
            evt.ws_sent_time = Date.now() * 1000;
            
            const payload = JSON.stringify(evt);
            for (const ws of clients) {
                if (ws.readyState === WebSocket.OPEN) {
                    ws.send(payload);
                }
            }
        }
    }
}

module.exports = WebsocketService;
