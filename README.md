# pageturner

A self-hosted book tracking application with ratings, reading stats, and OpenLibrary integration.

## Overview

pageturner is a full-stack book logging app that lets you track books you've read, rate them, and view reading statistics. It integrates with [OpenLibrary](https://openlibrary.org) for book search and metadata lookup.

## Features

- Track books with ratings, start/finish dates, page counts, and personal notes
- Search OpenLibrary by title or author
- View reading statistics by genre, month, and average rating
- Contribution-style chart of pages read over time
- Commonplace book view for quotes and highlights
- Export books to CSV or JSON
- REST API for all data endpoints

## Architecture

```
pageturner/
├── backend/              # Go API server
│   ├── internal/
│   │   ├── api/          # HTTP handlers, middleware, routes (chi)
│   │   ├── db/           # sqlc queries, schema, migrations
│   │   ├── repository/   # Data access layer
│   │   └── types/        # Shared domain types
│   ├── migrations/       # PostgreSQL migrations (golang-migrate)
│   ├── scripts/          # OpenLibrary data import
│   └── sqlc.yaml         # sqlc code generation config
├── frontend/             # SvelteKit web app
│   └── web/
│       ├── src/routes/   # SvelteKit routes (books, commonplace)
│       └── e2e/          # Playwright tests
├── bruno/                # API test collection
└── docker-compose.yml    # Full stack orchestration
```

### Tech Stack

- **Backend**: Go, chi router, pgx/v5, sqlc, golang-migrate, slog
- **Frontend**: SvelteKit, TypeScript, Tailwind CSS, Better Auth, Drizzle ORM
- **Database**: PostgreSQL 16
- **Infrastructure**: Docker Compose, Playwright (e2e tests), Bruno (API tests)

## Quick Start

```bash
# Set required environment variables
export BETTER_AUTH_SECRET=$(openssl rand -hex 16)

# Build and start all services
docker-compose up --build
```

Access the application at http://localhost:5173

### Services

| Service    | Port | Description                          |
|------------|------|--------------------------------------|
| Frontend   | 5173 | SvelteKit web application            |
| Backend    | 8080 | Go API server                        |
| PostgreSQL | 5432 | Database                             |

### Configuration

The backend reads from `backend/config.yml` by default. Environment variables override config file values:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=booktracker
DB_PASSWORD=booktracker
DB_NAME=booktracker
SSL_MODE=disable
```

## Development

### Backend

```bash
cd backend
go run main.go          # Interactive prompt
go run main.go --migrate  # Run migrations
go run main.go S         # Start server
```

### Frontend

```bash
cd frontend/web
npm install
npm run dev
```

### Tests

```bash
# Backend tests
cd backend && go test ./...

# Frontend e2e tests
cd frontend/web && npx playwright test
```

## License

MIT