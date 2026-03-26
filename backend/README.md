# Backend API Reference

The backend is a Go REST + WebSocket API built with the **Gin** framework, **GORM** ORM, and **PostgreSQL**. It handles authentication, auction lifecycle management, real-time bidding via WebSocket, and a credit-based wallet system.

**Base URL (development):** `http://localhost:8080`

---

## Table of Contents

- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [User Endpoints](#user-endpoints)
  - [Auction Endpoints](#auction-endpoints)
  - [Wallet Endpoints](#wallet-endpoints)
  - [Admin Endpoints](#admin-endpoints)
  - [Upload Endpoint](#upload-endpoint)
  - [WebSocket](#websocket)
- [Data Models](#data-models)
- [Error Responses](#error-responses)
- [Business Logic](#business-logic)
  - [Bidding System](#bidding-system)
  - [Anti-Sniping](#anti-sniping)
  - [Auction Worker](#auction-worker)
  - [Wallet & Credit Flow](#wallet--credit-flow)

---

## Authentication

The API uses **JWT Bearer tokens**. Include the access token in every authenticated request:

```
Authorization: Bearer <access_token>
```

**Token lifetimes:**
- Access token: **15 minutes** (HS256, signed with `JWT_ACCESS_SECRET`)
- Refresh token: **7 days** (HS256, signed with `JWT_REFRESH_SECRET`)

When an access token expires, call `POST /users/refresh` with the stored refresh token to obtain a new access token without re-logging in. The refresh token is stored in the user record — logging out revokes it server-side.

For WebSocket connections, pass the token as a query parameter instead:
```
ws://localhost:8080/ws?token=<access_token>
```

---

## Endpoints

### User Endpoints

#### `POST /users/register`

Creates a new user account and an empty wallet in a single database transaction.

**Auth required:** No

**Request body:**
```json
{
  "username": "alex_hunter",
  "email": "alex@example.com",
  "password": "SecurePass123"
}
```

**Validation rules:**
- `username`: 3–20 characters, alphanumeric only
- `email`: valid email format
- `password`: 8–72 characters

**Success `201`:**
```json
{ "message": "user created successfully" }
```

**Errors:**
- `400` — validation failed (returns field-level errors)
- `409` — username or email already taken

---

#### `POST /users/login`

Authenticates the user. Returns an access token and a refresh token. The refresh token is stored server-side (single device — logging in on a new device invalidates the previous refresh token).

**Auth required:** No

**Request body:**
```json
{
  "email": "alex@example.com",
  "password": "SecurePass123"
}
```

**Success `200`:**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

**Errors:**
- `401` — invalid email or password

---

#### `POST /users/refresh`

Exchanges a valid refresh token for a new access token. The existing refresh token remains valid.

**Auth required:** No

**Request body:**
```json
{ "refresh_token": "eyJ..." }
```

**Success `200`:**
```json
{ "access_token": "eyJ..." }
```

**Errors:**
- `401` — refresh token invalid, expired, or does not match the stored token
- `403` — account has been banned

---

#### `POST /users/logout`

Revokes the user's refresh token server-side. The access token remains valid until it naturally expires (15 minutes).

**Auth required:** Yes

**Success `200`:**
```json
{ "message": "logged out successfully" }
```

---

#### `GET /users/me`

Returns the authenticated user's full profile, including their wallet.

**Auth required:** Yes

**Success `200`:**
```json
{
  "ID": 5,
  "Username": "alex_hunter",
  "Email": "alex@example.com",
  "Role": "USER",
  "IsActive": true,
  "Wallet": {
    "ID": 3,
    "UserID": 5,
    "Balance": 10000,
    "ReservedBalance": 500
  }
}
```

---

#### `PATCH /users/change-password`

Updates the authenticated user's password after verifying the current one.

**Auth required:** Yes

**Request body:**
```json
{
  "old_password": "SecurePass123",
  "new_password": "NewSecurePass456"
}
```

**Success `200`:**
```json
{ "message": "password updated successfully" }
```

**Errors:**
- `401` — old password incorrect

---

### Auction Endpoints

#### `GET /auctions`

Returns a paginated list of **ACTIVE** auctions only, ordered by creation date (newest first).

**Auth required:** No

**Query parameters:**

| Parameter | Default | Max | Description |
|---|---|---|---|
| `page` | `1` | — | Page number |
| `limit` | `20` | `100` | Items per page |

**Success `200`:**
```json
{
  "page": 1,
  "limit": 20,
  "auctions": [
    {
      "ID": 12,
      "Title": "Apple MacBook Pro 16\" M3 Max",
      "ImageURL": "https://...",
      "StartingPrice": 2500,
      "BidIncrement": 50,
      "CurrentHighestBid": 2850,
      "CurrentHighestBidderID": 7,
      "BidCount": 7,
      "Status": "ACTIVE",
      "StartTime": "2026-03-26T10:00:00Z",
      "EndTime": "2026-03-28T10:00:00Z",
      "CreatedBy": 1
    }
  ]
}
```

---

#### `GET /auctions/:id`

Returns a single auction by ID, regardless of status.

**Auth required:** No

**Success `200`:** Returns the full `Auction` object (same shape as above).

**Errors:**
- `404` — auction not found

---

#### `GET /auctions/:id/bids`

Returns the complete bid history for an auction, ordered by time descending. Each bid includes a nested `user` object with just the username (to avoid leaking email/password data).

**Auth required:** No

**Success `200`:**
```json
[
  {
    "ID": 88,
    "AuctionID": 12,
    "UserID": 7,
    "Amount": 2850,
    "CreatedAt": "2026-03-26T14:22:10Z",
    "user": { "ID": 7, "Username": "jake_torres" }
  }
]
```

---

#### `POST /auctions/:id/bid`

Places a bid on an active auction. This is the most logic-intensive endpoint — see [Bidding System](#bidding-system) for full details.

**Auth required:** Yes

**Request body:**
```json
{ "amount": 2900 }
```

**Success `201`:**
```json
{
  "message": "bid placed successfully",
  "bid": { "ID": 89, "AuctionID": 12, "UserID": 5, "Amount": 2900, "CreatedAt": "..." }
}
```

**Errors:**
- `400` — bid too low, auction not active, or invalid input
- `400` — `"sellers cannot bid on their own auctions"`
- `400` — `"insufficient credits"`
- `404` — auction not found

---

### Wallet Endpoints

Both wallet endpoints require authentication. The wallet tracks `Balance` (total credits) and `ReservedBalance` (credits locked in active bids). **Available = Balance − ReservedBalance**.

#### `GET /users/wallet`

**Auth required:** Yes

**Success `200`:**
```json
{
  "ID": 3,
  "UserID": 5,
  "Balance": 10000,
  "ReservedBalance": 2900
}
```

---

#### `GET /users/wallet/transactions`

Returns paginated credit transaction history for the authenticated user, newest first.

**Auth required:** Yes

**Query parameters:** `page` (default `1`), `limit` (default `20`, max `100`)

**Transaction types:**

| Type | Meaning |
|---|---|
| `ADMIN_ASSIGN` | Credits added by an admin |
| `BID_RESERVE` | Credits locked when placing a bid |
| `BID_RELEASE` | Credits returned when outbid |
| `AUCTION_WIN` | Credits permanently deducted on winning |

**Success `200`:**
```json
{
  "page": 1,
  "limit": 20,
  "transactions": [
    {
      "ID": 42,
      "UserID": 5,
      "Amount": 2900,
      "Type": "BID_RESERVE",
      "Reference": "auction_12",
      "CreatedAt": "2026-03-26T14:22:10Z"
    }
  ]
}
```

---

### Admin Endpoints

All `/admin/*` routes require:
1. A valid access token (`Authorization: Bearer ...`)
2. The user's role to be `ADMIN` — enforced by `RoleRequired("ADMIN")` middleware

---

#### `POST /admin/auctions`

Creates a new auction. The backend automatically sets the status:
- **`ACTIVE`** if `start_time` is in the past or within 5 seconds of now
- **`SCHEDULED`** if `start_time` is in the future (the auction worker will activate it)

**Request body:**
```json
{
  "title": "Rolex Submariner",
  "description": "Full set with box and papers.",
  "image_url": "https://...",
  "starting_price": 12000,
  "bid_increment": 200,
  "start_time": "2026-03-27T10:00:00Z",
  "end_time": "2026-03-29T10:00:00Z"
}
```

**Validation:**
- `title`: min 3 characters, required
- `starting_price`: must be > 0
- `bid_increment`: must be > 0 and ≤ `starting_price`
- `end_time` must be at least 1 minute after `start_time`
- Both times must be in RFC3339 format

**Success `201`:** Returns the full created `Auction` object.

---

#### `GET /admin/auctions`

Returns **all** auctions regardless of status, ordered by creation date. Used by the admin dashboard to show a complete picture including SCHEDULED, ENDED, and CANCELLED auctions.

**Success `200`:**
```json
{ "auctions": [ /* all auctions */ ] }
```

---

#### `POST /admin/auctions/:id/end`

Force-closes an active auction immediately, regardless of its scheduled end time. Triggers full credit settlement (winner pays, previous bidder already refunded during bidding). Uses `AdminForceCloseAuction` which bypasses the time-check guard used by the background worker.

**Success `200`:**
```json
{ "message": "auction force-closed successfully" }
```

---

#### `POST /admin/auctions/:id/cancel`

Cancels an auction. If there is a current highest bidder, their reserved credits are released back to their available balance. The auction status is set to `CANCELLED`.

**Success `200`:**
```json
{ "message": "auction cancelled and credits refunded" }
```

---

#### `GET /admin/users`

Returns all registered users, each with their wallet preloaded.

**Success `200`:** Array of user objects including nested `Wallet`.

---

#### `PATCH /admin/users/:user_id/credits`

Adds credits to a user's wallet. Creates a `ADMIN_ASSIGN` credit transaction for the audit trail. Uses a database transaction so the wallet update and the transaction record are atomic.

**Request body:**
```json
{ "amount": 5000 }
```

**Success `200`:**
```json
{ "message": "credits assigned successfully" }
```

---

#### `PATCH /admin/promote/:user_id`

Sets a user's `Role` to `ADMIN`.

**Success `200`:**
```json
{ "message": "user promoted to admin" }
```

---

#### `PATCH /admin/users/:user_id/ban`

Sets `IsActive = false` on the user. Banned users cannot log in — the refresh endpoint checks `IsActive` before issuing a new access token.

**Success `200`:**
```json
{ "message": "user banned successfully" }
```

---

### Upload Endpoint

#### `POST /upload`

Accepts a `multipart/form-data` request with a single `image` field. Validates the file extension and size, generates a unique filename (`{timestamp}_{8-char-hex}{ext}`), and saves it to the `./uploads/` directory. Only used in local development — production uploads go directly to Cloudinary from the browser.

**Auth required:** Yes

**Content-Type:** `multipart/form-data`

**Field:** `image` (file)

**Allowed types:** `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`

**Max size:** 10 MB

**Success `200`:**
```json
{ "url": "http://localhost:8080/uploads/1743000000000_a3f9b2c1.jpg" }
```

Uploaded files are served as static assets at `GET /uploads/:filename`.

---

### WebSocket

#### `GET /ws`

Upgrades an HTTP connection to a WebSocket. Requires the access token as a query parameter.

```
ws://localhost:8080/ws?token=<access_token>
```

**Client → Server messages:**

```json
{ "type": "JOIN_AUCTION",  "auction_id": 12 }
{ "type": "LEAVE_AUCTION", "auction_id": 12 }
```

Joining an auction room subscribes the client to all real-time events for that auction.

**Server → Client messages:**

| `type` | When sent | Key fields |
|---|---|---|
| `BID_UPDATE` | A new bid is placed | `auction_id`, `amount`, `bidder_id` |
| `OUTBID` | The previous highest bidder is outbid | `auction_id`, `amount`, `bidder_id` |
| `AUCTION_EXTENDED` | Anti-snipe extension triggered | `auction_id`, `end_time` |
| `AUCTION_END` | Auction finalized by worker or admin | `auction_id` |
| `AUCTION_START` | Scheduled auction becomes active | `auction_id` |

---

## Data Models

### User
```
ID              uint
Username        string   (unique, 3–20 alphanum)
Email           string   (unique, lowercase)
PasswordHash    string   (bcrypt, cost 10)
Role            string   ("USER" | "ADMIN")
IsActive        bool     (false = banned)
RefreshToken    string   (current valid refresh token)
Wallet          Wallet   (preloaded on /users/me and /admin/users)
```

### Wallet
```
ID               uint
UserID           uint    (unique FK → User)
Balance          int64   (total credits; never goes negative)
ReservedBalance  int64   (credits locked in active bids)
```

Available balance = `Balance − ReservedBalance`

### Auction
```
ID                      uint
Title                   string
Description             string
ImageURL                string
StartingPrice           int64
BidIncrement            int64
CurrentHighestBid       int64   (= StartingPrice initially)
CurrentHighestBidderID  *uint   (nil until first bid)
BidCount                int64
ExtensionCount          int     (anti-snipe counter, max 10)
Status                  string  ("ACTIVE" | "SCHEDULED" | "ENDED" | "CANCELLED")
StartTime               time.Time
EndTime                 time.Time
CreatedBy               uint    (FK → User)
```

### Bid
```
ID         uint
AuctionID  uint
UserID     uint
Amount     int64
CreatedAt  time.Time
User       User   (preloaded with username only on /auctions/:id/bids)
```

### CreditTransaction
```
ID         uint
UserID     uint
Amount     int64
Type       string   ("ADMIN_ASSIGN" | "BID_RESERVE" | "BID_RELEASE" | "AUCTION_WIN")
Reference  string   (e.g. "auction_12", "auction_12_refund")
CreatedAt  time.Time
```

---

## Error Responses

All error responses follow the same shape:

```json
{ "error": "human-readable message" }
```

Validation errors return a map of field-level messages:

```json
{ "errors": { "email": "must be a valid email", "password": "min length is 8" } }
```

---

## Business Logic

### Bidding System

`POST /auctions/:id/bid` runs inside a single database transaction with a `SELECT … FOR UPDATE` lock on the auction row, preventing race conditions when two users bid simultaneously.

The following checks happen in order:
1. Auction exists and is `ACTIVE`
2. Current time is between `StartTime` and `EndTime`
3. The bidder is not the auction creator
4. The bid amount meets the minimum:
   - If no bids yet: `amount >= StartingPrice`
   - Otherwise: `amount >= CurrentHighestBid + BidIncrement`
5. The bidder has sufficient available credits

**Credit flow on a successful bid:**

- **New bidder:** `ReservedBalance += bidAmount`. The previous highest bidder's full reserved amount is released: `ReservedBalance -= prevAmount`, `Balance += prevAmount` (logged as `BID_RELEASE`).
- **Raising your own bid:** Only the *difference* is reserved: `ReservedBalance += (newAmount − currentAmount)`.

All credit changes are recorded as `CreditTransaction` rows within the same transaction.

---

### Anti-Sniping

If a bid is placed within the final **30 seconds** of an auction, the end time is extended by **30 seconds**. This prevents last-second sniping. Extensions are capped at **10** per auction (stored in `ExtensionCount`). After each extension, a `AUCTION_EXTENDED` WebSocket message is broadcast to all clients watching that auction so the countdown timer updates in real time.

---

### Auction Worker

A background goroutine polls the database every **2 seconds** for two conditions:

1. **Scheduled → Active:** Any auction with `status = 'SCHEDULED'` whose `start_time <= now` is transitioned to `ACTIVE`, and an `AUCTION_START` WebSocket event is broadcast.

2. **Active → Ended:** Any auction with `status = 'ACTIVE'` whose `end_time <= now` is finalized:
   - Status set to `ENDED`
   - If there is a winner: `ReservedBalance -= winAmount`, `Balance -= winAmount` (net deduction from wallet), logged as `AUCTION_WIN`
   - An `AUCTION_END` WebSocket event is broadcast to all connected clients in the auction room

The worker's `FinalizeAuction` function includes a time guard (`time.Now().Before(auction.EndTime)`) to prevent early finalization in race conditions. Admin force-close uses a separate `AdminForceCloseAuction` function that bypasses this guard.

---

### Wallet & Credit Flow

All wallet mutations use database transactions with `SELECT … FOR UPDATE` locks to prevent double-spending in concurrent bid scenarios.

```
User action          Balance    ReservedBalance    Available
─────────────────────────────────────────────────────────────
Initial (admin)      +10,000           0           10,000
Place $500 bid       10,000          +500           9,500
Outbid (refunded)    10,000          -500          10,000
Win $800 auction     -800            -800           9,200
```

The wallet never permits `Balance - ReservedBalance < 0` — a bid that would overdraw available credits is rejected with `"insufficient credits"`.
