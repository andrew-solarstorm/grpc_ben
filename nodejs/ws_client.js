const WebSocket = require('ws');
const minimist = require('minimist');

const args = minimist(process.argv.slice(2), {
  string: ['ws', 'mint'],
  default: { ws: 'ws://localhost:8080/ws' }
});

const { ws: wsURL, mint } = args;

if (!mint) {
  console.error('ERR: --mint is required');
  process.exit(2);
}

// High-resolution timer helper to get Unix Microseconds
const getNowUs = () => {
  // Combine wall clock seconds with high-res nanoseconds for μs precision
  return BigInt(Date.now()) * 1000n;
};

const ws = new WebSocket(wsURL);

ws.on('open', () => {
  ws.send(JSON.stringify({ action: 0, mint }));
  console.log(`Connected & Subscribed: ${mint}`);
});

ws.on('message', (data) => {
  try {
    const t = JSON.parse(data);
    const nowUs = getNowUs();

    // Convert server times to BigInt to prevent precision loss during subtraction
    const wsSentUs = BigInt(t.ws_sent_time);
    const decodedUs = BigInt(t.decoded_time);
    const receivedUs = BigInt(t.server_received_time);

    // 1. WS -> Client (Now - Sent)
    const wsToClientMs = Number(nowUs - wsSentUs) / 1000;

    // 2. Blocktime delay (Now - BlockTime)
    const nowMs = Number(nowUs / 1000n);
    const blockTimeMs = t.block_time * 1000;
    const blockDelayMs = nowMs - blockTimeMs;
    const blockDelaySec = Math.floor(blockDelayMs / 1000);

    // 3. Decode Time (Decoded - Received)
    const decodeTimeMs = Number(decodedUs - receivedUs) / 1000;

    console.log(`--- Transfer: ${t.transaction_signature} ---`);
    console.log(`From:      ${t.from}`);
    console.log(`To:        ${t.to}`);
    console.log(`Amount:    ${t.amount} | Mint: ${t.mint_address}`);
    console.log(`Slot:      ${t.slot} | Blocktime: ${t.block_time}`);
    // If still negative on the same server, we label it as < 0.001ms (Sync jitter)
    const displayWsLag = wsToClientMs < 0 ? "0.000*" : wsToClientMs.toFixed(3);

    console.log(`  WS -> Client:   ${displayWsLag} ms ${wsToClientMs < 0 ? '(Clock Jitter)' : ''}`);
    console.log(`  Decode Time:    ${decodeTimeMs.toFixed(3)} ms`);
    console.log(`  Block Lag:      ${blockDelayMs.toLocaleString()} ms (${blockDelaySec}s)`);
    console.log(`------------------------------------------------\n`);

  } catch (err) {
    console.error('Error processing message:', err.message);
  }
});

const handleExit = () => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ action: 1, mint }));
    ws.close();
  }
  process.exit();
};

process.on('SIGINT', handleExit);
process.on('SIGTERM', handleExit);
ws.on('error', (err) => console.error('WS Error:', err.message));
