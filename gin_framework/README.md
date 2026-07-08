# Cashier App (Gin Framework)

A Go + Gin cashier service with a small web client to manage bill inventory, calculate change, and expose runtime cashier metrics.

## Features

- Manage bill/coin inventory with denomination validation
- Calculate change using available inventory
- Persist inventory updates after dispensing change
- Expose register metrics endpoint
- Serve frontend client from the same backend server

## Tech Stack

- Go 1.26+
- Gin
- PostgreSQL
- sqlc
- goose migrations
- Testify (integration assertions)

## Project Structure

- `cmd/main.go`: app entrypoint and route registration
- `internal/handlers`: HTTP handlers
- `internal/database`: sqlc-generated DB access code
- `internal/config`: environment config loader
- `sql/schema`: goose migrations
- `sql/queries`: SQL queries for sqlc
- `client`: static frontend files

## Prerequisites

- Go installed
- PostgreSQL running locally (or reachable DB URL)
- goose installed (for migrations)

Install goose (if needed):

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Environment Variables

Create a `.env` file in the project root:

```env
DB_URL=postgres://carlosinfante@localhost:5432/cashier?sslmode=disable
PORT=8080
```

Notes:

- `DB_URL` is preferred, `DATABASE_URL` is also supported by config loader.
- `PORT` is required.

## Database Setup

From the project root:

```bash
cd sql/schema
goose postgres "postgres://carlosinfante@localhost:5432/cashier?sslmode=disable" up
```

Current migrations:

- `001_cashier.sql`: creates `bills`
- `002_quantity_non_negative.sql`: adds non-negative quantity constraint

## Run the App

From the project root:

```bash
go run cmd/main.go
```

Server starts on `http://localhost:8080` (or your configured `PORT`).

## Available Routes

### Health

- `GET /health`

Response example:

```json
{
  "status": "ok",
  "version": "1.0",
  "message": "Cashier API is healthy"
}
```

### Bills

- `GET /api/bills`
- `POST /api/bills`

Create or update bill quantity request body:

```json
[
  {
    "denomination": 20,
    "quantity": 5
  }
]
```

Rules:

- Quantity must be `>= 0`
- Denomination must be one of:
  - `100, 50, 20, 10, 5, 1, 0.5, 0.2, 0.1, 0.05, 0.01`

### Change

- `POST /api/change`

Request body example:

```json
{
  "amount_due": 13,
  "amount_paid": 20
}
```

Response example:

```json
[
  { "text": "1 x €5 = €5" },
  { "text": "1 x €2 = €2" }
]
```

Notes:

- Returns 400 if paid amount is lower than due amount.
- Returns 400 if register cannot provide exact change.
- Successfully calculated change is deducted from inventory.

### Metrics

- `GET /metrics`

Response example:

```json
{
  "denominations_count": 4,
  "total_bills_count": 18,
  "total_amount_cents": 11600,
  "total_amount_euro": 116
}
```

### Frontend

Served statically by backend:

- `GET /`
- `GET /index.html`
- `GET /app.js`
- `GET /styles.css`

Open in browser:

- `http://localhost:8080`

## Test

Run all tests:

```bash
go test ./...
```

Run change integration test only:

```bash
TEST_DB_URL=postgres://carlosinfante@localhost:5432/cashier?sslmode=disable go test ./internal/handlers -run TestGetChangeRoute_Integration -v
```

If `TEST_DB_URL` is not set, the integration test falls back to `DB_URL`.
