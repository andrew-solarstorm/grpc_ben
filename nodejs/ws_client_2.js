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

let highestBlockLagMs = -Infinity;
let lowestBlockLagMs = Infinity;
let messagesReceived = 0;

ws.on('open', () => {
  ws.send(JSON.stringify({ action: 0, mint }));
  console.log(`Connected & Subscribed: ${mint}`);
});

ws.on('message', (data) => {
  try {
    const t = JSON.parse(data);
    const nowUs = getNowUs();

    // 2. Blocktime delay (Now - BlockTime)
    const nowMs = Number(nowUs / 1000n);
    const blockTimeMs = t.block_time * 1000;
    const blockDelayMs = nowMs - blockTimeMs;
    const blockDelaySec = Math.floor(blockDelayMs / 1000);

    if (blockDelayMs >= 0) {
      if (blockDelayMs > highestBlockLagMs) highestBlockLagMs = blockDelayMs;
      if (blockDelayMs < lowestBlockLagMs) lowestBlockLagMs = blockDelayMs;
    }
    messagesReceived++;


    console.log(`--- Transfer: ${t.transaction_signature} ---`);
    console.log(`From:      ${t.from}`);
    console.log(`To:        ${t.to}`);
    console.log(`Amount:    ${t.amount} | Mint: ${t.mint_address}`);
    console.log(`Slot:      ${t.slot} | Blocktime: ${t.block_time}`);
    console.log(`Block Lag:          ${blockDelayMs.toLocaleString()} ms (${blockDelaySec}s)`);
    console.log(`------------------------------------------------\n`);

  } catch (err) {
    console.error('Error processing message:', err.message);
  }
});

const handleExit = () => {
  if (messagesReceived > 0) {
    console.log(`\n=== Session Summary ===`);
    const displayHighest = highestBlockLagMs === -Infinity ? 'N/A' : highestBlockLagMs.toLocaleString();
    const displayLowest = lowestBlockLagMs === Infinity ? 'N/A' : lowestBlockLagMs.toLocaleString();
    console.log(`Highest Block Lag: ${displayHighest} ms`);
    console.log(`Lowest Block Lag:  ${displayLowest} ms`);
    console.log(`=======================\n`);
  }
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ action: 1, mint }));
    ws.close();
  }
  process.exit();
};

process.on('SIGINT', handleExit);
process.on('SIGTERM', handleExit);
ws.on('error', (err) => console.error('WS Error:', err.message));
