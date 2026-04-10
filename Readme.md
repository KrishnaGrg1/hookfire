
# Hookfire

A webhook delivery engine built in Go. Receives events and reliably delivers them to subscriber URLs with automatic retries and exponential backoff.

## Why Go

| | Hookfire (Go) | Node.js equivalent |
|---|---|---|
| Memory per worker | ~2KB | ~1MB |
| 1000 concurrent deliveries | ~10MB RAM | ~500MB RAM |
| Deployment | Single 8MB binary | 400MB with node_modules |

## Architecture

```
Client -> POST /api/v1/events
					 |
					 +-- Save to PostgreSQL
					 +-- Push to Redis queue
										|
							 Worker pool (goroutines)
										|
							POST to subscriber URLs
										|
					 success -> log
					 failure -> retry with backoff
```

## Stack

- Go + Chi - HTTP server
- PostgreSQL + sqlc - persistent storage
- Redis - job queue
- Goose - migrations

## Prerequisites

- Go 1.25+
- PostgreSQL running locally
- Goose CLI installed
- sqlc CLI installed

## Environment Variables

Copy `.env.example` to `.env` and update values for your local database.

```bash
cp .env.example .env
```

Example:

```env
PORT=8080

GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://postgres:<password>@localhost:5432/hookfire?sslmode=disable
GOOSE_MIGRATION_DIR=./migrations
```

Note: The app currently connects using `GOOSE_DBSTRING`.

## Run the Application

```bash
# Run migrations
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING=DATABASE_URL \
goose -dir migrations up

# Start server
go run ./cmd
```

## Database Migrations (Goose)

1. Create a new migration file:

```bash
goose -s create create_<table_or_change_name> sql
```

2. Open the generated file in `migrations/` and write your SQL under the Goose sections (`-- +goose Up` and `-- +goose Down`).

3. Apply all pending migrations to the database:

```bash
goose up
```

4. Verify migration status:

```bash
goose status
```

Example migration file content:

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS applications (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	api_key TEXT UNIQUE NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS applications;
```

## SQL Queries and Code Generation (sqlc)

1. Add SQL queries in `internal/db/query/*.sql`
2. Use sqlc query annotations, for example:

```sql
-- name: CreateApplication :one
INSERT INTO applications(name, api_key)
VALUES($1, $2)
RETURNING id, name, api_key, created_at;
```

3. Generate Go code:

```bash
sqlc generate
```

Generated files are written to `internal/db/sqlc`.

Server starts on `PORT` and exposes:

- `GET /` -> `Hello, Hookfire`

## API

```bash
# Create an application
curl -X POST http://localhost:8080/api/v1/applications \
	-H "Content-Type: application/json" \
	-d '{"name": "my app"}'

# Register a subscriber endpoint
curl -X POST http://localhost:8080/api/v1/endpoints \
	-H "Authorization: Bearer hf_yourkey" \
	-H "Content-Type: application/json" \
	-d '{"url": "https://yourapp.com/webhooks"}'

# Send an event
curl -X POST http://localhost:8080/api/v1/events \
	-H "Authorization: Bearer hf_yourkey" \
	-H "Content-Type: application/json" \
	-d '{"event_type": "payment.success", "payload": {"amount": 500}}'
```


## Retry Policy

| Attempt | Delay |
|---|---|
| 1 | 10s |
| 2 | 20s |
| 3 | 40s |
| 4 | 80s |
| 5 | Dead letter |

## Docker

Docker images run migrations automatically on startup.

### Option A: Docker Compose (local dev)

```bash
docker compose up --build
```

### Option B: Docker run (build or pull)
1. Build the image (Dockerfile is only for building)
```bash
docker build -t <image_name>:<tag> .
```
If you skipped the build, pull from Docker Hub instead:
```bash
docker pull krishnagrg/hookfire:latest
```

2. Create network
```bash
docker network create hooknet
```
3. Create volume
```bash
docker volume create pgdata
``` 
4. Run PostgreSQL
```bash
docker run -d \
  --name postgres \
  --network hooknet \
  -v pgdata:/var/lib/postgresql/data \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=pass123 \
  -e POSTGRES_DB=hookfire \
  -p 5432:5432 \
  postgres:16-alpine
```

5. Run Redis
```bash
docker run -d \
  --name redis \
  --network hooknet \
  -p 6379:6379 \
  redis:7-alpine
```
6. Run Hookfire
```bash
docker run --rm -p 8080:8080 \
  --network hooknet \
  -e PORT=8080 \
  -e GOOSE_DRIVER=postgres \
  -e GOOSE_DBSTRING="postgres://postgres:pass123@postgres:5432/hookfire?sslmode=disable" \
  -e GOOSE_MIGRATION_DIR=/app/migrations \
  -e REDIS_URL="redis://redis:6379" \
  -e WORKER_COUNT=10 \
  <image_name>:<tag>
```

## Common Errors

- Goose parse error about unfinished query:
	- Ensure each SQL statement ends with `;`
- sqlc metadata error:
	- Query annotation must use `-- name:` (with a space)
- sqlc `:one` error without returning clause:
	- Add a `RETURNING ...` clause to insert/update queries marked as `:one`