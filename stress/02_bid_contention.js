/**
 * TEST 2: Bid contention — many users hammering ONE auction
 *
 * All VUs target the same auction. This is the worst-case Redis lock
 * contention scenario: only one bid can be processed at a time, so all
 * others get "auction is busy, please try again".
 *
 * What this measures:
 *   - Lock contention behavior under heavy load
 *   - "auction busy" rate vs successful bid rate
 *   - Latency of the lock-acquire → reject path (fast path)
 *   - Whether the server crashes or degrades gracefully
 *
 * Note: 400 "bid too low" and "auction is busy" are EXPECTED responses —
 * they are NOT errors. Only 5xx are true failures.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const BASE_URL = 'http://localhost:8080';

// Single low-price auction for maximum contention.
// auction 529: starting=2500, increment=100, no existing bids — clean slate.
// auction 631: starting=250, increment=10, 0 existing bids — clean slate.
const TARGET_AUCTION = 631;

const successfulBids = new Counter('successful_bids');
const busyRejections = new Counter('busy_rejections');
const lowBidRejections = new Counter('low_bid_rejections');
const serverErrors = new Rate('server_errors');
const bidLatency = new Trend('bid_latency');

// Seed users: alex_hunter, sarah_m, jake_torres, ...@example.com / Password123
const USERS = [
  'alex_hunter', 'sarah_m', 'jake_torres', 'emma_vance', 'ryan_cole',
  'olivia_park', 'noah_west', 'chloe_reed', 'ethan_shaw', 'ava_brooks',
  'liam_foster', 'mia_grant', 'mason_hayes', 'sophia_lin', 'lucas_ford',
  'isabella_wu', 'aiden_price', 'ella_morgan', 'jackson_bell', 'grace_kim',
];

export const options = {
  stages: [
    { duration: '10s', target: 10 },
    { duration: '30s', target: 20 },  // max 20 — one per user account
    { duration: '20s', target: 20 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    server_errors: ['rate<0.01'],        // <1% server errors (5xx)
    bid_latency: ['p(95)<2000'],         // 95th percentile under 2s (includes DB tx)
  },
};

export function setup() {
  const tokens = {};
  for (const username of USERS) {
    const res = http.post(
      `${BASE_URL}/users/login`,
      JSON.stringify({ email: `${username}@example.com`, password: 'Password123' }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    if (res.status === 200) {
      tokens[username] = res.json('access_token');
    } else {
      console.warn(`Login failed for ${username}: ${res.status} ${res.body}`);
    }
  }
  console.log(`Logged in ${Object.keys(tokens).length} users`);
  return { tokens };
}

export default function (data) {
  // Assign each VU a deterministic user so no two VUs share an account
  const vuIndex = (__VU - 1) % USERS.length;
  const username = USERS[vuIndex];
  const token = data.tokens[username];

  if (!token) {
    console.error(`No token for ${username}`);
    return;
  }

  // Get current auction state to know minimum bid
  const stateRes = http.get(`${BASE_URL}/auctions/${TARGET_AUCTION}`);
  if (stateRes.status !== 200) return;

  const auction = stateRes.json();
  const currentBid = auction.CurrentHighestBid || 0;
  const startingPrice = auction.StartingPrice;
  const increment = auction.BidIncrement;

  // Compute a bid that would be valid IF we win the race
  const minBid = currentBid === 0 ? startingPrice : currentBid + increment;
  // Add a random extra increment so all VUs don't bid the same amount
  const bidAmount = minBid + Math.floor(Math.random() * 5) * increment;

  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/auctions/${TARGET_AUCTION}/bid`,
    JSON.stringify({ amount: bidAmount }),
    { headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    }},
  );
  bidLatency.add(Date.now() - start);

  const is5xx = res.status >= 500;
  serverErrors.add(is5xx);

  if (res.status === 201) {
    successfulBids.add(1);
  } else if (res.status === 400) {
    const body = res.json();
    if (body && body.error && body.error.includes('busy')) {
      busyRejections.add(1);
    } else {
      lowBidRejections.add(1);
    }
  }

  check(res, {
    'not a server error': (r) => r.status < 500,
  });

  sleep(0.1);
}
