/**
 * TEST 3: Parallel bid throughput — each user on a different auction
 *
 * Each VU is assigned its own dedicated auction. There is zero Redis lock
 * contention between VUs. This measures the server's TRUE maximum bid
 * throughput when load is spread evenly across auctions.
 *
 * This is the test whose output belongs on your resume:
 *   "X bids/sec sustained across Y concurrent users with P95 latency Zms"
 *
 * Architecture note: Because each auction has its own Redis lock, N VUs
 * bidding on N different auctions achieve N× the throughput of N VUs
 * all bidding on the same auction. This is the parallelism benefit of
 * the per-auction lock design.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

const BASE_URL = 'http://localhost:8080';

// 20 active auctions (one per user), all with affordable starting prices.
// Auction IDs from latest seed. Users have 10,000 credits each.
// 20 active auctions (one per user), all with affordable starting prices.
// Auction IDs from latest seed. Users have 10,000 credits each.
const AUCTION_ASSIGNMENTS = [
  { username: 'alex_hunter',  auctionID: 631, increment: 10  },
  { username: 'sarah_m',      auctionID: 639, increment: 25  },
  { username: 'jake_torres',  auctionID: 637, increment: 25  },
  { username: 'emma_vance',   auctionID: 632, increment: 25  },
  { username: 'ryan_cole',    auctionID: 638, increment: 25  },
  { username: 'olivia_park',  auctionID: 633, increment: 25  },
  { username: 'noah_west',    auctionID: 628, increment: 25  },
  { username: 'chloe_reed',   auctionID: 640, increment: 25  },
  { username: 'ethan_shaw',   auctionID: 629, increment: 25  },
  { username: 'ava_brooks',   auctionID: 626, increment: 25  },
  { username: 'liam_foster',  auctionID: 643, increment: 25  },
  { username: 'mia_grant',    auctionID: 625, increment: 50  },
  { username: 'mason_hayes',  auctionID: 636, increment: 50  },
  { username: 'sophia_lin',   auctionID: 630, increment: 25  },
  { username: 'lucas_ford',   auctionID: 627, increment: 50  },
  { username: 'isabella_wu',  auctionID: 644, increment: 50  },
  { username: 'aiden_price',  auctionID: 635, increment: 50  },
  { username: 'ella_morgan',  auctionID: 642, increment: 50  },
  { username: 'jackson_bell', auctionID: 634, increment: 50  },
  { username: 'grace_kim',    auctionID: 641, increment: 50  },
];

const successfulBids = new Counter('successful_bids');
const insufficientCredits = new Counter('insufficient_credits');
const serverErrors = new Rate('server_errors');
const bidLatency = new Trend('bid_latency');
const fullCycleLatency = new Trend('full_cycle_latency');

export const options = {
  scenarios: {
    parallel_bidders: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 20 },
        { duration: '40s', target: 20 },
        { duration: '10s', target: 0 },
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    server_errors: ['rate<0.01'],
    bid_latency: ['p(95)<1000'],
    successful_bids: ['count>50'],
  },
};

export function setup() {
  const tokens = {};
  for (const assignment of AUCTION_ASSIGNMENTS) {
    const res = http.post(
      `${BASE_URL}/users/login`,
      JSON.stringify({ email: `${assignment.username}@example.com`, password: 'Password123' }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    if (res.status === 200) {
      tokens[assignment.username] = res.json('access_token');
    } else {
      console.warn(`Login failed for ${assignment.username}: ${res.status}`);
    }
  }
  console.log(`Logged in ${Object.keys(tokens).length}/${AUCTION_ASSIGNMENTS.length} users`);
  return { tokens };
}

export default function (data) {
  // Each VU maps to exactly one user/auction pair
  const idx = (__VU - 1) % AUCTION_ASSIGNMENTS.length;
  const { username, auctionID, increment } = AUCTION_ASSIGNMENTS[idx];
  const token = data.tokens[username];
  if (!token) return;

  const cycleStart = Date.now();

  // Fetch current auction state
  const stateRes = http.get(`${BASE_URL}/auctions/${auctionID}`);
  if (stateRes.status !== 200) return;

  const auction = stateRes.json();
  if (auction.Status !== 'ACTIVE') return;

  const currentBid = auction.CurrentHighestBid || 0;
  const startingPrice = auction.StartingPrice;
  const isTopBidder = auction.CurrentHighestBidderID != null;

  // Compute minimum valid bid for this VU
  let minBid;
  if (currentBid === 0) {
    minBid = startingPrice;
  } else {
    minBid = currentBid + increment;
  }

  // Place the bid
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/auctions/${auctionID}/bid`,
    JSON.stringify({ amount: minBid }),
    { headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    }},
  );
  bidLatency.add(Date.now() - start);
  fullCycleLatency.add(Date.now() - cycleStart);

  const is5xx = res.status >= 500;
  serverErrors.add(is5xx);

  if (res.status === 201) {
    successfulBids.add(1);
  } else if (res.status === 400) {
    const body = res.json();
    if (body && body.error && body.error.includes('insufficient')) {
      insufficientCredits.add(1);
    }
  } else if (is5xx) {
    console.error(`5xx on auction ${auctionID}: ${res.body}`);
  }

  check(res, {
    'bid: no server error': (r) => r.status < 500,
  });

  sleep(0.2);
}
