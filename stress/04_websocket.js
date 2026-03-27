/**
 * TEST 4: WebSocket connection capacity
 *
 * Establishes many concurrent persistent WebSocket connections. Each VU:
 *   1. Connects to ws://localhost:8080/ws?token=<jwt>
 *   2. Joins an auction room (JOIN_AUCTION message)
 *   3. Stays connected for the duration
 *   4. Counts incoming messages (BID_UPDATE, etc.)
 *
 * Run this test WHILE test 03 is running to simulate real load:
 * observers (WS) + bidders (HTTP) happening simultaneously.
 *
 * What this measures:
 *   - Maximum concurrent WebSocket connections
 *   - Connection establishment latency (p95)
 *   - Message delivery rate under connection load
 *   - Whether the hub goroutine becomes a bottleneck
 */

import ws from 'k6/ws';
import { check, sleep } from 'k6';
import http from 'k6/http';
import { Counter, Gauge, Rate, Trend } from 'k6/metrics';

const BASE_URL = 'http://localhost:8080';
const WS_URL   = 'ws://localhost:8080';

const AUCTION_IDS = [631, 639, 637, 632, 638, 633, 628, 640, 629, 626, 643, 625, 636, 630, 627, 644, 635, 642, 634, 641];

const USERS = [
  'alex_hunter', 'sarah_m', 'jake_torres', 'emma_vance', 'ryan_cole',
  'olivia_park', 'noah_west', 'chloe_reed', 'ethan_shaw', 'ava_brooks',
  'liam_foster', 'mia_grant', 'mason_hayes', 'sophia_lin', 'lucas_ford',
  'isabella_wu', 'aiden_price', 'ella_morgan', 'jackson_bell', 'grace_kim',
];

const connectErrors    = new Rate('ws_connect_errors');
const messagesReceived = new Counter('ws_messages_received');
const activeConns      = new Gauge('ws_active_connections');
const connectLatency   = new Trend('ws_connect_latency');

export const options = {
  stages: [
    { duration: '10s', target: 50 },
    { duration: '20s', target: 200 },
    { duration: '30s', target: 500 },
    { duration: '20s', target: 1000 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    ws_connect_errors: ['rate<0.05'],          // <5% connection failures
    ws_connect_latency: ['p(95)<2000'],        // connect in under 2s
  },
};

export function setup() {
  // Login all 20 users and store their tokens
  const tokens = [];
  for (const username of USERS) {
    const res = http.post(
      `${BASE_URL}/users/login`,
      JSON.stringify({ email: `${username}@example.com`, password: 'Password123' }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    if (res.status === 200) {
      tokens.push(res.json('access_token'));
    }
  }
  console.log(`Got ${tokens.length} tokens for WS test`);
  return { tokens };
}

export default function (data) {
  if (!data.tokens || data.tokens.length === 0) {
    console.error('No tokens available');
    return;
  }

  // Cycle through available tokens (many VUs will share tokens — that's fine
  // for WS testing; we're measuring connection capacity, not user isolation)
  const token = data.tokens[(__VU - 1) % data.tokens.length];
  const auctionID = AUCTION_IDS[(__VU - 1) % AUCTION_IDS.length];

  const connectStart = Date.now();

  const res = ws.connect(
    `${WS_URL}/ws?token=${token}`,
    { tags: { auction: String(auctionID) } },
    function (socket) {
      connectLatency.add(Date.now() - connectStart);
      activeConns.add(1);

      socket.on('open', () => {
        // Join a specific auction room
        socket.send(JSON.stringify({ type: 'JOIN_AUCTION', auction_id: auctionID }));
      });

      socket.on('message', (data) => {
        messagesReceived.add(1);
        // Verify message is valid JSON with expected shape
        try {
          const msg = JSON.parse(data);
          check(msg, {
            'message has type': (m) => typeof m.type === 'string',
            'message has auction_id': (m) => typeof m.auction_id === 'number',
          });
        } catch (e) {
          // ignore parse errors in counters
        }
      });

      socket.on('error', (e) => {
        connectErrors.add(1);
        console.error(`WS error VU${__VU}: ${e}`);
      });

      socket.on('close', () => {
        activeConns.add(-1);
      });

      // Hold connection open for 20 seconds
      socket.setTimeout(() => {
        socket.send(JSON.stringify({ type: 'LEAVE_AUCTION', auction_id: auctionID }));
        socket.close();
      }, 20000);
    },
  );

  const connected = check(res, {
    'WS connection established (101)': (r) => r && r.status === 101,
  });
  if (!connected) {
    connectErrors.add(1);
  }

  // Small sleep between VU iterations so connections don't all expire at once
  sleep(1);
}
