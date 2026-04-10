#!/bin/sh
set -e

echo "Running migrations..."

max_attempts=10
attempt=1

while [ "$attempt" -le "$max_attempts" ]; do
  if /app/goose -dir /app/migrations postgres "$GOOSE_DBSTRING" up; then
    echo "Migrations successful"
    break
  fi

  echo "Migration failed, retrying in 2s..."
  attempt=$((attempt + 1))
  sleep 2
done

if [ "$attempt" -gt "$max_attempts" ]; then
  echo "Migration failed after retries"
  exit 1
fi

echo "Starting app..."
exec /app/hookfire