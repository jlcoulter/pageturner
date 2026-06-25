#!/bin/sh
set -e

echo "Running database migrations..."
cd /app

# Run goose migrations
./booktracker --migrate

echo "Starting server..."
exec ./booktracker S