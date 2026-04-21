const express = require('express');
const { createServer } = require('http');
const { WebSocketServer } = require('ws');
const minimist = require('minimist');

const { WebsocketService } = require('./ws');
const { Decoder } = require('./decoder');
const { SystemClock } = require('./sysclock');
const { TxIngestService } = require('./tx_ingest');
const { BlockBuilder } = require('./block_ingestion');
const { Buffer } = require('./buffer');

async function main() {
    let endpoint, token, commitmentStr, httpAddr;

    const argv = minimist(process.argv.slice(2), {
        string: ['endpoint', 'token', 'commitment', 'http'],
        default: {
            endpoint: 'http://localhost:10000',
            token: '',
            commitment: 'PROCESSED',
            http: ':::8080'
        }
    });

    endpoint = argv.endpoint;
    token = argv.token;
    commitmentStr = argv.commitment;
    httpAddr = argv.http;

    if (httpAddr) {
        if (!httpAddr.includes(':')) {
            httpAddr = ':::' + httpAddr;
        } else if (httpAddr.startsWith(':')) {
            httpAddr = '::' + httpAddr;
        }
    }

    console.log(`ENDPOINT: ${endpoint} TOKEN: ${token} CommitmentLevel: ${commitmentStr} HTTP: ${httpAddr}`);

    if (!token) {
        console.log("ERR: token is not set");
        return;
    }

    // Initialize exactly like Go version
    const sysClock = new SystemClock();
    await sysClock.subscribe(endpoint, token, commitmentStr);

    const wsSvc = new WebsocketService();
    const dec = new Decoder(wsSvc);

    const blockBuilder = new BlockBuilder(dec);
    
    const buff = new Buffer(sysClock, blockBuilder);

    const ingest = new TxIngestService(buff);
    await ingest.Subscribe(endpoint, token, commitmentStr);

    // REST and Websockets setup
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

    const portArg = httpAddr.substring(httpAddr.lastIndexOf(':') + 1);
    const hostArg = httpAddr.substring(0, httpAddr.lastIndexOf(':'));

    server.listen(parseInt(portArg), hostArg, () => {
        console.log(`🚀 HTTP/WS Server listening on ${httpAddr}`);
    });

    server.on('error', (err) => {
        console.log(`http server error: ${err}`);
    });

    // Handle OS signals like `<-sigChan`
    const shutdown = () => {
        ingest.Close();
        sysClock.Close();
        
        server.close(() => {
            console.log("✅ block_pulling completed");
            process.exit(0);
        });
        
        setTimeout(() => process.exit(0), 5000); // hard stop like context.WithTimeout
    };

    process.on('SIGINT', shutdown);
    process.on('SIGTERM', shutdown);
}

main().catch(err => {
    console.error(err);
    process.exit(1);
});
