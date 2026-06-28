#!/bin/sh
set -e

echo "Running database migrations..."
cd /app

# Run goose migrations
./pageturner --migrate

echo "Starting server..."
exec ./pageturner S