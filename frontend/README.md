# Frontend Architecture

The frontend is a React 19 TypeScript SPA built with Vite 8. It communicates with the Go backend over HTTP (Axios) and WebSocket (native browser API). This document explains how every major piece is wired together.

---

## Table of Contents

- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Connecting to the Backend](#connecting-to-the-backend)
  - [Axios Client & JWT Interceptor](#axios-client--jwt-interceptor)
  - [API Modules](#api-modules)
  - [WebSocket Connection](#websocket-connection)
- [Authentication Flow](#authentication-flow)
- [Context Providers](#context-providers)
  - [AuthContext](#authcontext)
  - [SocketContext](#socketcontext)
- [Routing & Route Protection](#routing--route-protection)
- [Pages](#pages)
  - [Home](#home)
  - [Login & Register](#login--register)
  - [Auctions](#auctions)
  - [Auction Detail](#auction-detail)
  - [Wallet](#wallet)
  - [Admin](#admin)
- [Styling System](#styling-system)
- [Running the Frontend](#running-the-frontend)

---

## Tech Stack

| Package | Version | Purpose |
|---|---|---|
| React | 19 | UI rendering |
| TypeScript | 5 | Type safety |
| Vite | 8 | Dev server + bundler |
| React Router DOM | 7 | Client-side routing |
| Axios | 1 | HTTP client with interceptors |
| CSS Modules | — | Scoped component styles |

No component libraries (no MUI, no Tailwind). All UI is hand-crafted with CSS Modules and CSS custom properties.

---

## Project Structure

```
frontend/src/
│
├── api/                  # One file per backend resource
│   ├── client.ts         # Axios instance + JWT interceptor + refresh logic
│   ├── auth.ts           # register, login, logout, refresh, getMe
│   ├── auctions.ts       # getAuctions, getAuction, getBids, placeBid
│   ├── wallet.ts         # getWallet, getTransactions
│   ├── admin.ts          # createAuction, fetchUsers, assignCredits, …
│   └── upload.ts         # uploadImage (Cloudinary or local backend)
│
├── context/
│   ├── AuthContext.tsx   # User + wallet state; login/logout/refreshWallet
│   └── SocketContext.tsx # WebSocket lifecycle; joinAuction/subscribe
│
├── hooks/
│   └── useCountdown.ts   # Ticking countdown string + state ("warning"/"expired")
│
├── pages/
│   ├── Home.tsx
│   ├── Login.tsx
│   ├── Register.tsx
│   ├── Auctions.tsx
│   ├── AuctionDetail.tsx
│   ├── Wallet.tsx
│   └── Admin.tsx
│
├── components/
│   ├── Navbar.tsx
│   ├── ProtectedRoute.tsx
│   ├── AuctionCard.tsx
│   ├── Countdown.tsx
│   ├── StatusBadge.tsx
│   ├── Skeleton.tsx
│   ├── Pagination.tsx
│   └── admin/
│       ├── UserTable.tsx
│       ├── AdminAuctionTable.tsx
│       └── CreateAuctionForm.tsx
│
├── styles/
│   ├── variables.css     # Design tokens (colours, spacing, typography)
│   └── global.css        # Reset, body, layout helpers, shared form/button classes
│
├── types/
│   └── index.ts          # TypeScript interfaces for all API shapes
│
├── App.tsx               # Route declarations
└── main.tsx              # React root + context provider tree
```

---

## Connecting to the Backend

### Axios Client & JWT Interceptor

**`src/api/client.ts`** is the central HTTP client. Every API call in the app goes through this single Axios instance so that auth headers and token refresh are handled in one place.

**Request interceptor** — attaches the access token to every outgoing request:
```ts
config.headers.Authorization = `Bearer ${localStorage.getItem('access_token')}`
```

**Response interceptor** — handles token expiry transparently:
1. If a response returns `401` and the request hasn't already been retried, the interceptor tries to refresh the token silently.
2. It calls `POST /users/refresh` with the refresh token from `localStorage`.
3. On success, the new access token is stored and all queued requests (that arrived while the refresh was in flight) are retried with the new token.
4. On failure (refresh token also expired or revoked), `localStorage` is cleared and the user is redirected to `/login`.

```ts
// Simplified from client.ts
client.interceptors.response.use(
  (res) => res,
  async (error) => {
    if (error.response?.status !== 401 || original._retry) return Promise.reject(error)
    original._retry = true
    // ... refresh logic with queuing for concurrent requests
  }
)
```

The queue pattern prevents multiple parallel 401s from each triggering their own refresh call — only the first one refreshes, the rest wait and then retry once the new token is available.

---

### API Modules

Each file in `src/api/` maps to a backend resource. All functions return Axios promises, so call sites use `.then()` or `await`.

| File | Functions |
|---|---|
| `auth.ts` | `loginUser`, `registerUser`, `logoutUser`, `refreshToken`, `getCurrentUser`, `changePassword` |
| `auctions.ts` | `fetchAuctions(page, limit)`, `fetchAuction(id)`, `fetchBids(id)`, `placeBid(id, amount)` |
| `wallet.ts` | `getWallet()`, `getTransactions(page, limit)` |
| `admin.ts` | `fetchUsers`, `assignCredits`, `promoteUser`, `banUser`, `getAdminAuctions`, `createAuction`, `endAuction`, `cancelAuction` |
| `upload.ts` | `uploadImage(file)` — see [Image Upload](#image-upload) below |

**TypeScript types** in `src/types/index.ts` mirror the backend Go structs exactly. Because the Go structs have no `json` tags, they serialize with **PascalCase** field names (`CurrentHighestBid`, not `current_highest_bid`). All TypeScript interfaces match this casing.

```ts
// Types use PascalCase to match Go struct serialization
export interface Auction {
  ID: number
  Title: string
  CurrentHighestBid: number
  BidCount: number
  Status: 'ACTIVE' | 'SCHEDULED' | 'ENDED' | 'CANCELLED'
  EndTime: string
  // ...
}
```

The exception is `WsMessage`, whose Go struct *does* have `json` tags, so it uses snake_case:
```ts
export interface WsMessage {
  type: string
  auction_id: number
  amount: number
  bidder_id: number
  end_time?: string
}
```

---

#### Image Upload

**`src/api/upload.ts`** abstracts over two upload destinations:

- **Production (Cloudinary):** If `VITE_CLOUDINARY_CLOUD_NAME` and `VITE_CLOUDINARY_UPLOAD_PRESET` env vars are set, images are uploaded directly from the browser to Cloudinary using an unsigned upload preset. No backend involvement — Cloudinary returns a CDN URL.
- **Development (local backend):** If the env vars are absent, the file is sent as `multipart/form-data` to `POST /upload` on the Go server, which saves it to `./uploads/` and returns a local URL.

This means the `CreateAuctionForm` component doesn't need to know where the image is going — it just calls `uploadImage(file)` and gets back a URL.

---

### WebSocket Connection

The WebSocket is managed entirely by `SocketContext`. It connects to `ws://localhost:8080/ws?token=<access_token>` using the native browser `WebSocket` API (no library needed).

**Lifecycle:**
1. Connects when the user logs in (`isAuthenticated` becomes `true`)
2. Disconnects and cleans up on logout
3. Reconnects automatically after a 3-second delay if the connection drops unexpectedly

**Message routing** uses a `Set` of handler functions stored in a `useRef` (so it doesn't trigger re-renders). Any component can subscribe to all messages and filter by `auction_id` itself.

```ts
// Subscribe pattern
const unsubscribe = subscribe((msg: WsMessage) => {
  if (msg.auction_id !== auctionId) return
  // handle msg.type
})
// Cleanup on unmount:
return () => unsubscribe()
```

**Joining a room** — to receive events for a specific auction, the client sends:
```json
{ "type": "JOIN_AUCTION", "auction_id": 12 }
```
And on unmount:
```json
{ "type": "LEAVE_AUCTION", "auction_id": 12 }
```

---

## Authentication Flow

The flow from cold start to authenticated session:

```
App mounts
    │
    ▼
AuthProvider reads localStorage
    │
    ├─ No access_token → setLoading(false) → show unauthenticated UI
    │
    └─ access_token found
           │
           ▼
       GET /users/me   ← Axios attaches token automatically
           │
           ├─ 200 → setUser(data), GET /users/wallet → setWallet(data)
           │
           └─ 401 → interceptor calls POST /users/refresh
                       │
                       ├─ success → retry /users/me with new token
                       └─ failure → clear localStorage, redirect /login
```

**Login action** (`AuthContext.login`):
1. Calls `POST /users/login`
2. Stores `access_token` and `refresh_token` in `localStorage`
3. Immediately calls `GET /users/me` and `GET /users/wallet` to hydrate state
4. React state updates cascade to all consumers

**Logout action** (`AuthContext.logout`):
1. Calls `POST /users/logout` (revokes server-side refresh token)
2. Clears `localStorage`
3. Sets `user` and `wallet` to `null`
4. Navigates to `/login`

---

## Context Providers

The provider tree in `main.tsx`:

```tsx
<BrowserRouter>
  <AuthProvider>
    <SocketProvider>
      <App />
    </SocketProvider>
  </AuthProvider>
</BrowserRouter>
```

`SocketProvider` is nested inside `AuthProvider` so it can read `isAuthenticated` to know when to connect/disconnect.

---

### AuthContext

**Exposes:**

| Value | Type | Description |
|---|---|---|
| `user` | `User \| null` | Authenticated user profile |
| `wallet` | `Wallet \| null` | Current wallet balances |
| `isAuthenticated` | `boolean` | Shorthand for `!!user` |
| `isAdmin` | `boolean` | Shorthand for `user?.Role === 'ADMIN'` |
| `loading` | `boolean` | True while initial session is being restored |
| `login(email, pw)` | `async fn` | Full login flow |
| `logout()` | `async fn` | Full logout flow |
| `refreshWallet()` | `async fn` | Re-fetches wallet from backend (called after placing a bid) |

`refreshWallet` is called by `AuctionDetail` after a successful bid to update the navbar balance without requiring a page refresh.

---

### SocketContext

**Exposes:**

| Value | Type | Description |
|---|---|---|
| `joinAuction(id)` | `fn` | Sends `JOIN_AUCTION` to backend |
| `leaveAuction(id)` | `fn` | Sends `LEAVE_AUCTION` to backend |
| `subscribe(handler)` | `fn → unsub fn` | Registers a message handler; returns cleanup function |

The WebSocket instance is stored in a `useRef` — it never causes re-renders when the connection state changes. The `handlers` set is also a ref for the same reason. Both `joinAuction`, `leaveAuction`, and `subscribe` are wrapped in `useCallback` with no dependencies so they are stable references (safe to use in `useEffect` dependency arrays).

---

## Routing & Route Protection

**`src/App.tsx`** declares all routes using React Router DOM v7:

```
/              → Home (public)
/login         → Login (public)
/register      → Register (public)
/auctions      → Auctions grid (public)
/auctions/:id  → Auction detail + bidding (public to view, auth to bid)
/wallet        → Wallet (requires login)
/admin         → Admin dashboard (requires ADMIN role)
```

**`ProtectedRoute`** wraps any route that requires authentication or a specific role:

```tsx
<ProtectedRoute requiredRole="ADMIN">
  <Admin />
</ProtectedRoute>
```

It checks `AuthContext.loading` first — while the session is being restored from `localStorage`, it renders nothing (preventing a flash-redirect before we know if the user is logged in). Once loading is complete:
- If not authenticated → redirect to `/login`
- If authenticated but wrong role → redirect to `/`
- Otherwise → render children

---

## Pages

### Home

A landing page with a hero section and a "How it works" explainer. The hero links directly to `/auctions`. No data fetching.

---

### Login & Register

Standard controlled forms. Both use `AuthContext.login` / `registerUser` + navigate. Validation errors from the backend are displayed inline beneath the form. The register page redirects to `/login` on success.

---

### Auctions

Fetches `GET /auctions?page=N&limit=12` and renders a responsive grid of `AuctionCard` components. Pagination is handled by `Pagination` (prev/next buttons). While loading, skeleton placeholder cards are shown.

Each `AuctionCard` displays:
- Auction image (or a grey placeholder if no image)
- Title
- Current bid (or starting price if no bids)
- `StatusBadge` (ACTIVE / ENDED / SCHEDULED)
- Countdown timer via the `Countdown` component (only for ACTIVE)
- Number of bids

---

### Auction Detail

The most complex page. Layout is two-column on desktop (image + bid history on the left, bid panel on the right) and single-column on mobile.

**Data loading:**
Fetches `GET /auctions/:id` and `GET /auctions/:id/bids` in parallel using `Promise.all`. If the auction is not found, the user is redirected to `/auctions`.

**Real-time updates via WebSocket:**

```ts
// On mount
joinAuction(auctionId)
const unsub = subscribe(handleMessage)

// On unmount
leaveAuction(auctionId)
unsub()
```

The `handleMessage` function handles three server-sent event types:

| Event | Action |
|---|---|
| `BID_UPDATE` | Updates `CurrentHighestBid`, increments `BidCount`, prepends the new bid to the bid history list |
| `AUCTION_EXTENDED` | Updates `endTime` state (which feeds `useCountdown`), shows a 5-second "⚡ Extended by 30s" banner |
| `AUCTION_END` | Sets auction `Status` to `ENDED`, disabling the bid form |

**Bid placement:**
The bid form calls `POST /auctions/:id/bid`, then calls `refreshWallet()` on success to update the navbar balance. The minimum bid is calculated client-side and shown as the input placeholder (`≥ $2,900`).

**Status banners:**
Conditionally rendered based on the user's relationship to the auction:
- "You are currently winning" (green) / "You have been outbid" (amber)
- "You won this auction!" (on ENDED if user is winner)
- "Sign in to place a bid" (unauthenticated)
- "You created this auction" (creator)

---

### Wallet

Fetches `GET /users/wallet` and `GET /users/wallet/transactions`. Displays three balance cards (Available, Reserved, Total) and a paginated transaction list.

Each transaction type has a distinct colour dot and sign prefix:
- `ADMIN_ASSIGN` — green dot, `+`
- `BID_RESERVE` — amber dot, `−`
- `BID_RELEASE` — blue dot, `+`
- `AUCTION_WIN` — red dot, `−`

---

### Admin

Tabbed interface with two panels: **Users** and **Auctions**.

**Users tab** (`UserTable` component):
- Displays all users with username, email, role badge, active/banned status, and available wallet balance
- "Assign Credits" — inline number input + "Add" button per row; calls `PATCH /admin/users/:id/credits`
- "Ban" button — confirms with `window.confirm`, then calls `PATCH /admin/users/:id/ban`; hidden for admin accounts

**Auctions tab:**
- `CreateAuctionForm` — grid layout form with file upload. On file selection, the image is immediately uploaded (to Cloudinary or local backend) and a preview is shown. The resulting URL is sent with the auction creation request.
- `AdminAuctionTable` — shows all auctions (all statuses), with "Force End" and "Cancel" action buttons

---

## Styling System

**Design tokens** (`src/styles/variables.css`) define the entire colour palette and spacing scale as CSS custom properties:

```css
:root {
  --bg:          #0f0f0f;  /* page background */
  --surface:     #1a1a1a;  /* cards */
  --surface-2:   #222;     /* hover states */
  --surface-3:   #2a2a2a;  /* inputs */
  --border:      #2a2a2a;
  --text:        #f0f0f0;
  --text-muted:  #888;
  --accent:      #f59e0b;  /* amber — bids, CTAs */
  --green:       #22c55e;
  --red:         #ef4444;
  --radius:      8px;
}
```

**CSS Modules** (`.module.css`) are used for every component and page. Class names are locally scoped — there are no global class collisions. The global stylesheet (`global.css`) only defines:
- CSS reset
- Body/typography defaults
- Shared utility classes: `.container`, `.page`, `.btn`, `.btn-primary`, `.btn-danger`, `.btn-sm`, `.form-input`, `.form-label`, `.form-group`, `.error-text`, `.success-text`

**Path aliases** — Vite is configured with `@` pointing to `./src`:
```ts
// vite.config.ts
resolve: { alias: { '@': path.resolve(__dirname, './src') } }
```

This allows clean imports everywhere:
```ts
import { useAuth } from '@/context/AuthContext'
import { fetchAuction } from '@/api/auctions'
```

---

## Running the Frontend

```bash
cd frontend
npm install
npm run dev       # Start dev server at http://localhost:5173
npm run build     # Production build → dist/
npm run preview   # Preview the production build locally
```

Make sure the Go backend is running on port 8080 before starting the dev server.

**Environment variables** (create `frontend/.env.local`):
```env
# Only needed for production image uploads (Cloudinary)
VITE_CLOUDINARY_CLOUD_NAME=your_cloud_name
VITE_CLOUDINARY_UPLOAD_PRESET=your_unsigned_preset
```

Leave these blank in local development — images will upload to the local backend instead.
