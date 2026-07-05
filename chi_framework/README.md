# Cashier API (Chi + PostgreSQL)

Simple cashier backend and web client to:
- manage bill inventory
- calculate change
- expose health and metrics endpoints

## Tech Stack

- Go 1.26+
- Chi router
- PostgreSQL
- pgx
- sqlc-generated queries
- Static frontend in client/

## Project Structure

- cmd/: application entrypoint and HTTP router
- internal/bills/: bills service and HTTP handlers
- internal/database/: sqlc-generated DB access layer
- sql/schema/: database schema
- sql/queries/: SQL queries used by sqlc
- client/: static web client served by the API

## Prerequisites

- Go installed
- PostgreSQL running locally on port 5432
- Database named cashier

## Environment Variables

Create a .env file in the project root:

DB_URL=postgres://carlosinfante@localhost:5432/cashier?sslmode=disable
PORT=8080

## Database Setup

Create the database and table:

```bash
createdb cashier
psql -d cashier -f sql/schema/001_cashier.sql
```

## Run the App

From the project root:

```bash
go run cmd/*.go
```

Server starts on:
- http://localhost:8080

Frontend client:
- http://localhost:8080/

## API Endpoints

- GET /health
- GET /metrics
- GET /api/bills
- POST /api/bills
- POST /api/change

### Example: Add or Update Bills

```bash
curl -X POST http://localhost:8080/api/bills \
  -H "Content-Type: application/json" \
  -d '[{"denomination":20,"quantity":5},{"denomination":5,"quantity":10}]'
```

Allowed denominations:
- 100, 50, 20, 10, 5, 1, 0.50, 0.20, 0.10, 0.05, 0.01

### Example: List Bills

```bash
curl http://localhost:8080/api/bills
```

### Example: Calculate Change

```bash
curl -X POST http://localhost:8080/api/change \
  -H "Content-Type: application/json" \
  -d '{"amount_due":10,"amount_paid":15}'
```

Expected response example:

```json
[
  {"text":"1 x €5 = €5"}
]
```

## Run Tests

Run all tests:

```bash
go test ./...
```

Run bills integration test only:

```bash
go test ./internal/bills -run TestGetChangeRoute_Integration_Success -v
```

## Notes

- The API stores denominations internally in cents to avoid floating-point precision issues.
- The client only shows available denominations (quantity > 0).
