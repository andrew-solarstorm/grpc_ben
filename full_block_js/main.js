const express = require('express');
const { createServer } = require('http');
const { WebSocketServer } = require('ws');
const minimist = require('minimist');

const { WebsocketService } = require('./ws');
const { Decoder } = require('./decoder');
const { BlockIngestionService } = require('./block_ingestion');

async function main() {
    const argv = minimist(process.argv.slice(2), {
        string: ['endpoint', 'token', 'commitment', 'http'],
        default: {
            endpoint: 'http://localhost:10000',
            token: '',
            commitment: 'PROCESSED',
            http: '0.0.0.0:8080'
        }
    });

    let httpAddr = argv.http;
    if (httpAddr) {
        if (!httpAddr.includes(':')) {
            httpAddr = '0.0.0.0:' + httpAddr;
        } else if (httpAddr.startsWith(':')) {
            httpAddr = '0.0.0.0' + httpAddr;
        }
    }

    console.log(`ENDPOINT: ${argv.endpoint} TOKEN: ${argv.token} CommitmentLevel: ${argv.commitment} HTTP: ${httpAddr}`);

    if (!argv.token) {
        console.log("ERR: token is not set");
        return;
    }

    const wsSvc = new WebsocketService();
    const dec = new Decoder(wsSvc);
    const ingest = new BlockIngestionService(dec);

    await ingest.Subscribe(argv.endpoint, argv.token, argv.commitment);

    // Express server
    const app = express();
    const server = createServer(app);
    const wss = new WebSocketServer({ noServer: true });

    app.get('/healthz', (req, res) => {
        res.send('ok');
    });

    server.on('upgrade', (request, socket, head) => {
        if (request.url === '/ws') {
            wss.handleUpgrade(request, socket, head, (ws) => {
                wsSvc.handleConnection(ws);
            });
        } else {
            socket.destroy();
        }
    });

    const [host, port] = httpAddr.split(':');
    
    server.listen(parseInt(port), host, () => {
        console.log(`HTTP/WS Server listening on ${httpAddr}`);
    });

    // Graceful shutdown
    const shutdown = () => {
        console.log("Shutting down...");
        ingest.Close();
        server.close(() => {
            console.log("✅ block_pulling completed");
            process.exit(0);
        });
        setTimeout(() => process.exit(1), 5000); // hard exit after 5s
    };

    process.on('SIGINT', shutdown);
    process.on('SIGTERM', shutdown);
}

main().catch(err => {
    console.error(err);
    process.exit(1);
});
