
# Hookfire

Hookfire is a Go backend project using:

- PostgreSQL
- Goose for database migrations
- sqlc for type-safe query code generation
- Chi for HTTP routing

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

## Run the Application

```bash
go run ./cmd
```

Server starts on `PORT` and exposes:

- `GET /` -> `Hello, Hookfire`

## Common Errors

- Goose parse error about unfinished query:
	- Ensure each SQL statement ends with `;`
- sqlc metadata error:
	- Query annotation must use `-- name:` (with a space)
- sqlc `:one` error without returning clause:
	- Add a `RETURNING ...` clause to insert/update queries marked as `:one`