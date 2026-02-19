# Expensify

A personal cashflow tracking app with charts. Log income and expenses, organize them by category, and visualize your spending over time.

## Features

- **Google OAuth login** — sign in with your Google account, no passwords
- **Cashflow entries** — log inflows (income) and outflows (expenses) with an amount, date, category, and optional description
- **Categories** — 12 built-in default categories (Food, Transport, Shopping, etc.) plus the ability to create custom ones with a custom icon and color
- **Charts**
  - Monthly bar chart showing inflow vs. outflow side by side
  - Spending-by-category pie chart with a color-coded legend
  - Period navigation: default view is the trailing 12 months; step back through calendar years with prev/next buttons
  - Summary stat cards: Total Inflow, Total Outflow, Net Balance
- **Pagination** — transaction list is paginated (20 per page)
- **Edit & delete** — update or remove any transaction; custom categories can be deleted (blocked if any transactions reference them)
- **Responsive** — works on desktop and mobile

## Tech stack

| Layer | Tech |
|---|---|
| Frontend | React 18, TypeScript, Vite, React Query, Recharts, Axios |
| Backend | Go 1.22, chi router, MongoDB |
| Auth | Google OAuth 2.0 + server-side sessions (HttpOnly cookies) |

## Project structure

```
expensify/
├── backend/
│   ├── cmd/server/          # main entrypoint
│   └── internal/
│       ├── api/             # HTTP handlers + router
│       ├── config/          # env-based config
│       ├── db/              # MongoDB repositories + seed data
│       ├── middleware/       # session auth middleware
│       ├── models/          # data models
│       ├── services/        # business logic
│       └── testutil/        # mock repositories for unit tests
└── frontend/
    └── src/
        ├── api/             # Axios API client functions
        ├── components/      # React UI components
        ├── hooks/           # React Query hooks
        ├── pages/           # Page-level components
        └── types/           # TypeScript types
```

## Prerequisites

- Go 1.22+
- Node.js 18+
- MongoDB (local or Atlas)
- A Google Cloud project with OAuth 2.0 credentials

## Running locally

### 1. Clone the repo

```bash
git clone <repo-url>
cd expensify
```

### 2. Set up Google OAuth credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com/) → APIs & Services → Credentials
2. Create an **OAuth 2.0 Client ID** (Web application)
3. Add `http://localhost:8080/auth/google/callback` as an authorized redirect URI

### 3. Configure the backend

```bash
cd backend
cp .env.example .env
```

Edit `.env` and fill in your values:

```env
MONGO_URI=mongodb://localhost:27017
DB_NAME=expensify

GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

SESSION_SECRET=at-least-32-random-characters
FRONTEND_URL=http://localhost:5173
PORT=8080
SECURE_COOKIES=false
```

### 4. Run the backend

```bash
cd backend
go run ./cmd/server
# Server starts on :8080
```

The server seeds the 12 default categories into MongoDB on first run automatically.

### 5. Configure the frontend

```bash
cd frontend
cp .env.example .env.local
```

For local development, leave `VITE_API_BASE_URL` empty — Vite proxies `/api` and `/auth` to the backend automatically.

### 6. Run the frontend

```bash
cd frontend
npm install
npm start
# App opens at http://localhost:5173
```

## Running tests

### Backend unit tests

```bash
cd backend
go test ./internal/services/...
```

### Backend integration tests (requires a running MongoDB)

```bash
cd backend
TEST_MONGO_URI=mongodb://localhost:27017 go test -tags integration ./internal/db/...
```

## Deploying to production

The app is designed to run with the frontend and backend on separate origins (e.g. `expensify.example.com` and `expensify-backend.example.com`).

### Backend env changes for production

```env
FRONTEND_URL=https://expensify.example.com
GOOGLE_REDIRECT_URL=https://expensify-backend.example.com/auth/google/callback
SECURE_COOKIES=true
SESSION_SECRET=<strong-random-secret>
```

### Frontend env changes for production

```env
VITE_API_BASE_URL=https://expensify-backend.example.com
```

Build the frontend:

```bash
cd frontend
npm run build
# Static files are output to frontend/dist/
```

Serve `frontend/dist/` with any static host (Nginx, Caddy, Vercel, etc.) and run the Go binary on your server.

## API reference

All API routes except the auth endpoints require a valid session cookie.

### Auth

| Method | Path | Description |
|---|---|---|
| `GET` | `/auth/google` | Redirect to Google login |
| `GET` | `/auth/google/callback` | OAuth callback, sets session cookie |
| `GET` | `/auth/me` | Returns the current user |
| `POST` | `/auth/logout` | Clears the session cookie |

### Categories

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/categories` | List all categories (defaults + custom) |
| `POST` | `/api/categories` | Create a custom category |
| `DELETE` | `/api/categories/:id` | Delete a custom category (blocked if transactions exist) |

### Transactions

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/transactions?page=1` | Paginated transaction list (20 per page) |
| `POST` | `/api/transactions` | Create a transaction |
| `PUT` | `/api/transactions/:id` | Update a transaction |
| `DELETE` | `/api/transactions/:id` | Delete a transaction |
| `GET` | `/api/cashflow/summary?months=12` | Aggregated monthly totals + category totals |
| `GET` | `/api/cashflow/summary?year=2025` | Same but for a specific calendar year |
