# Cashier App (No Framework)

A lightweight cashier API + static web client built with Go and PostgreSQL.

## Features

- Manage bill inventory by denomination
- Calculate change using available bills
- Deduct dispensed bills from inventory after change is calculated
- Simple browser client served by the same Go server

## Tech Stack

- Go (net/http)
- PostgreSQL
- sqlc-generated DB layer
- Static client (HTML, CSS, JavaScript)

## Project Structure

- `main.go`: server bootstrap and routes
- `handler_bills_create_update.go`: create/update bills
- `handler_bills_get.go`: list current bills
- `handler_get_change.go`: calculate and apply change
- `internal/database/`: sqlc generated queries and models
- `sql/schema/001_cashier.sql`: schema migration
- `client/`: static frontend

## Requirements

- Go 1.26+
- PostgreSQL

## Environment Variables

Create a `.env` file in this folder:

```env
DB_URL=postgres://USER:PASSWORD@localhost:5432/cashier_db?sslmode=disable
PORT=8080
```

## Database Setup

Create your database and run the schema in `sql/schema/001_cashier.sql`.

Example with `psql`:

```bash
createdb cashier_db
psql -d cashier_db -f sql/schema/001_cashier.sql
```

## Run

```bash
go run .
```

Server starts at:

- `http://localhost:8080`

## API Routes

### Health

- `GET /health`

Response example:

```json
{
  "status": "ok",
  "message": "server is running"
}
```

### Get Bills

- `GET /api/bills`

Response example:

```json
[
  {
    "id": "uuid",
    "created_at": "2026-06-30 10:00:00 +0000 UTC",
    "updated_at": "2026-06-30 10:00:00 +0000 UTC",
    "denomination": 20,
    "quantity": 5
  }
]
```

### Create/Update Bills

- `POST /api/bills`

Request body (array):

```json
[
  {
    "denomination": 20,
    "quantity": 10
  }
]
```

Behavior:

- If denomination does not exist, it creates a bill row
- If denomination exists, it replaces quantity with the provided value

### Calculate Change

- `POST /api/change`

Request body:

```json
{
  "amount_due": 18.75,
  "amount_paid": 50
}
```

Response example:

```json
[
  {
    "text": "1 x €20 = 20"
  },
  {
    "text": "1 x €10 = 10"
  },
  {
    "text": "1 x €1 = 1"
  },
  {
    "text": "1 x €0.25 = 0.25"
  }
]
```

Behavior:

- Validates enough money paid
- Validates enough total cash in store
- Computes change from available denominations
- Updates DB by subtracting dispensed quantities

## Client Notes

The static UI is served from `client/` and includes:

- Add/Update Bills form
- Calculate Change form
- Done button flow to clear fields/result and refresh current bills

## Common Troubleshooting

### `go tidy: unknown command`

Use:

```bash
go mod tidy
```

### Server fails to start

Check:

- `.env` exists
- `DB_URL` is valid
- `PORT` is set
- PostgreSQL is running

### Build check

```bash
go build ./...
```
