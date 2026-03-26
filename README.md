# Auction Platform

A full-stack real-time auction platform built with a Go/Gin REST + WebSocket backend and a React 19 TypeScript frontend. Users browse live auctions, place bids, and watch the competition unfold in real time. Administrators manage the entire platform — creating auctions, assigning credits, and moderating users — from a dedicated dashboard.

---

## Table of Contents

- [Features](#features)
- [User Roles](#user-roles)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation & Setup](#installation--setup)
- [Seed Data](#seed-data)
- [Documentation](#documentation)

---

## Features

- **Real-time bidding** via WebSocket — every bid, outbid notification, and auction extension appears instantly across all connected tabs
- **Anti-sniping** — a bid placed in the final 30 seconds extends the auction by 30 seconds (capped at 10 extensions)
- **Credit wallet** — a virtual credit system with full transaction history (reserves, releases, wins, admin assignments)
- **JWT authentication** — 15-minute access tokens with 7-day refresh tokens; silent token rotation on expiry
- **Role-based access control** — USER and ADMIN roles enforced at the middleware level
- **Image uploads** — direct upload to Cloudinary in production; local filesystem fallback in development
- **Paginated auction grid** — browse active auctions with page controls
- **Responsive dark UI** — CSS Modules + CSS custom properties, no component library dependency

---

## User Roles

### USER (Regular Bidder)

A standard account created through the registration page.

| Capability | Details |
|---|---|
| **Browse auctions** | View all active auctions with current bids and countdown timers |
| **Auction detail** | See full item description, image, bid history, and live bidding panel |
| **Place bids** | Bid on any auction you did not create; credits are reserved immediately |
| **Real-time updates** | Receive live bid updates, outbid alerts, and anti-snipe extension notices via WebSocket |
| **Wallet** | View available balance, reserved credits, and a full paginated transaction history |
| **Account** | Register, log in, log out, change password |

**Credit mechanics for regular users:**
- When you place a bid, the full bid amount is *reserved* from your wallet (unavailable for other bids)
- If you are outbid, your reserved amount is *released* back to your available balance instantly
- If you win an auction, the reserved amount is *deducted permanently* and logged as `AUCTION_WIN`
- Your wallet shows three figures: **Available** (spendable), **Reserved** (held in active bids), and **Total**

---

### ADMIN

Admins have all USER capabilities plus access to the `/admin` dashboard.

| Capability | Details |
|---|---|
| **Create auctions** | Set title, description, image, starting price, bid increment, start time, and end time |
| **View all auctions** | See every auction regardless of status (ACTIVE, SCHEDULED, ENDED, CANCELLED) |
| **Force-end an auction** | Immediately close any active auction and settle credits to the winner |
| **Cancel an auction** | Cancel an auction at any time; the highest bidder's reserved credits are refunded |
| **View all users** | See every registered user with their wallet balance |
| **Assign credits** | Add any amount of credits to any user's wallet |
| **Promote users** | Grant ADMIN role to a regular user |
| **Ban users** | Disable a user account, preventing future logins |

The default admin account is created automatically when the backend starts:
- **Email:** `admin@auction.com`
- **Password:** `admin123`

> Change this password immediately in any non-development environment.

---

## Tech Stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.25, Gin, GORM, PostgreSQL, Gorilla WebSocket, JWT (golang-jwt/jwt v5), bcrypt |
| **Frontend** | React 19, TypeScript, Vite 8, React Router DOM 7, Axios, CSS Modules |
| **Database** | PostgreSQL 14+ |
| **Image storage** | Cloudinary (production) / local filesystem (development) |

---

## Project Structure

```
auction/
├── backend/          # Go REST + WebSocket API
│   ├── cmd/
│   │   ├── main.go   # Server entry point
│   │   └── seed/     # Database seed script
│   ├── controllers/  # HTTP request handlers
│   ├── services/     # Business logic (bidding, wallet, WebSocket)
│   ├── models/       # GORM database models
│   ├── routes/       # Route registration
│   ├── middleware/   # Auth + RBAC middleware
│   ├── db/           # Database connection + migrations
│   └── utils/        # JWT helpers, validation
│
└── frontend/         # React TypeScript SPA
    └── src/
        ├── api/      # Axios API functions
        ├── context/  # AuthContext, SocketContext
        ├── pages/    # Full-page components
        ├── components/
        ├── hooks/
        ├── styles/   # Global CSS + design tokens
        └── types/    # TypeScript interfaces
```

---

## Prerequisites

Make sure the following are installed before proceeding:

| Tool | Version | Download |
|---|---|---|
| **Go** | 1.21+ | https://go.dev/dl/ |
| **Node.js** | 18+ | https://nodejs.org/ |
| **PostgreSQL** | 14+ | https://www.postgresql.org/download/ |
| **Git** | any | https://git-scm.com/ |

---

## Installation & Setup

### 1. Clone the repository

```bash
git clone <repository-url>
cd auction
```

### 2. Create the PostgreSQL database

Open `psql` or any PostgreSQL client and run:

```sql
CREATE DATABASE auction;
```

### 3. Configure the backend environment

Create `backend/.env` (copy from the example below):

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=auction

# JWT secrets — use long random strings in production
JWT_ACCESS_SECRET=replace_with_a_long_random_string
JWT_REFRESH_SECRET=replace_with_another_long_random_string
```

### 4. Install backend dependencies and start the server

```bash
cd backend
go mod download
go run ./cmd/main.go
```

The server starts on **http://localhost:8080**. On first run it:
- Runs all database migrations automatically
- Creates the default admin account (`admin@auction.com` / `admin123`)

### 5. Install frontend dependencies

```bash
cd ../frontend
npm install
```

### 6. (Optional) Configure image uploads for production

For local development, uploaded images are saved to `backend/uploads/` and served at `http://localhost:8080/uploads/`. For production (Render + Vercel), create `frontend/.env.local`:

```env
VITE_CLOUDINARY_CLOUD_NAME=your_cloud_name
VITE_CLOUDINARY_UPLOAD_PRESET=your_unsigned_preset_name
```

Sign up at [cloudinary.com](https://cloudinary.com), then go to **Settings → Upload → Add upload preset** (set signing mode to **Unsigned**).

### 7. Start the frontend

```bash
npm run dev
```

The app is now running at **http://localhost:5173**.

---

## Seed Data

To populate the database with 20 demo users and 100 realistic auctions (electronics, watches, art, sneakers, jewelry, and more):

```bash
cd backend
go run ./cmd/seed/main.go
```

All demo users have the password `Password123`. To wipe and re-seed:

```bash
go run ./cmd/seed/main.go --reset
```

---

## Documentation

- [Backend API Reference](backend/README.md) — every endpoint, request/response format, auth requirements, and business logic
- [Frontend Architecture](frontend/README.md) — how the backend is connected, how authentication works, real-time implementation, and page-by-page breakdown
