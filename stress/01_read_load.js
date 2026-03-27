/**
 * TEST 1: Read throughput under load
 *
 * Simulates many users browsing the auction list and viewing individual
 * auction pages. No auth required. Measures raw HTTP read performance —
 * the upper bound on what the server can serve.
 *
 * Targets:
 *   GET /auctions         (paginated list)
 *   GET /auctions/:id     (detail page)
 *   GET /auctions/:id/bids (bid history)
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = 'http://localhost:8080';

// Sample of active auction IDs from seed data
const AUCTION_IDS = [631, 639, 637, 632, 638, 633, 628, 640, 629, 626, 643, 625, 636, 630, 627, 644, 635, 642, 634, 641];

const errorRate = new Rate('errors');
const listLatency = new Trend('list_latency');
const detailLatency = new Trend('detail_latency');

export const options = {
  stages: [
    { duration: '15s', target: 50 },   // ramp up
    { duration: '30s', target: 200 },  // sustained load
    { duration: '30s', target: 500 },  // peak load
    { duration: '15s', target: 0 },    // ramp down
  ],
  thresholds: {
    errors: ['rate<0.01'],                    // <1% errors
    http_req_duration: ['p(95)<300'],         // 95th percentile under 300ms
    list_latency: ['p(95)<300'],
    detail_latency: ['p(95)<200'],
  },
};

export default function () {
  const action = Math.random();

  if (action < 0.4) {
    // 40% — browse auction list (random page)
    const page = Math.ceil(Math.random() * 3);
    const start = Date.now();
    const res = http.get(`${BASE_URL}/auctions?page=${page}`);
    listLatency.add(Date.now() - start);

    const ok = check(res, {
      'list: status 200': (r) => r.status === 200,
      'list: has auctions key': (r) => r.json('auctions') !== null,
    });
    errorRate.add(!ok);

  } else if (action < 0.85) {
    // 45% — view auction detail
    const id = AUCTION_IDS[Math.floor(Math.random() * AUCTION_IDS.length)];
    const start = Date.now();
    const res = http.get(`${BASE_URL}/auctions/${id}`);
    detailLatency.add(Date.now() - start);

    const ok = check(res, {
      'detail: status 200': (r) => r.status === 200,
      'detail: has ID': (r) => r.json('ID') > 0,
    });
    errorRate.add(!ok);

  } else {
    // 15% — view bid history
    const id = AUCTION_IDS[Math.floor(Math.random() * AUCTION_IDS.length)];
    const res = http.get(`${BASE_URL}/auctions/${id}/bids`);

    const ok = check(res, {
      'bids: status 200': (r) => r.status === 200,
    });
    errorRate.add(!ok);
  }

  sleep(0.05); // 50ms think time — aggressive but realistic for a SPA
}
