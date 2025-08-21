#!/bin/sh
set +e

DB_HOST=${DB_HOST:-postgres}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-alan}
DB_PASSWORD=${DB_PASSWORD:-2005}
DB_NAME=${DB_NAME:-orders_db}
PG_DSN=${PG_DSN:-postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}}

MAX_RETRIES=30
RETRY=0

echo "Waiting for Postgres at ${DB_HOST}:${DB_PORT} and applying migrations..."
while [ $RETRY -lt $MAX_RETRIES ]; do
  echo "pwd=$(pwd)"
  echo "listing /root/internal/migrations"
  ls -la /root/internal/migrations || ls -la ./internal/migrations || true
  /usr/local/bin/migrate -path file:///root/internal/migrations -database "${PG_DSN}?sslmode=disable" up
  RC=$?
  if [ $RC -eq 0 ]; then
    echo "Migrations applied successfully"
    break
  fi
  echo "migrate returned code=$RC, retry=$RETRY"
  RETRY=$((RETRY+1))
  sleep 1
done

if [ $RETRY -ge $MAX_RETRIES ]; then
  echo "Warning: migrations did not succeed after ${MAX_RETRIES} attempts, continuing to start the app"
fi

exec ./main
