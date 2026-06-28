# pageturner

A self-hosted book tracking application with ratings, reading stats, and OpenLibrary integration.

## Features

- Track books with ratings (0–10), start/finish dates, page counts, and personal notes
- Search OpenLibrary by title or author
- Commonplace book view for reading notes and highlights
- REST API for all data endpoints
- Docker Compose for one-command deployment

## Tech Stack

- **Backend**: Go, chi, pgx/v5, sqlc, goose, slog
- **Frontend**: SvelteKit, TypeScript, Tailwind CSS, Better Auth
- **Database**: PostgreSQL 16 (app data), SQLite (auth)
- **Testing**: Playwright (e2e), Bruno (API)

## Quick Start

```bash
# Generate an auth secret
export BETTER_AUTH_SECRET=$(openssl rand -hex 16)

# Build and start all services
docker compose up --build
```

Access the application at http://localhost:5173

### Services

| Service    | Port | Description               |
|------------|------|---------------------------|
| Frontend   | 5173 | SvelteKit web application |
| Backend    | 8080 | Go API server             |
| PostgreSQL | 5432 | Database                  |

### Configuration

Backend configuration via environment variables (overrides `backend/config.yml`):

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=pageturner
DB_PASSWORD=pageturner
DB_NAME=pageturner
SSL_MODE=disable
```

Frontend:

```bash
PUBLIC_API_URL=http://localhost:8080  # Backend API URL
BETTER_AUTH_SECRET=                    # Auth signing key
ORIGIN=http://localhost:5173          # SvelteKit origin
```

## Development

### Backend

```bash
cd backend
go run main.go              # Interactive prompt
go run main.go --migrate    # Run migrations only
go run main.go S            # Start server directly
```

### Frontend

```bash
cd frontend/web
npm install
npm run dev
```

### Tests

```bash
# Backend
cd backend && go test ./...

# Frontend e2e
cd frontend/web && npx playwright test
```

## Docker Commands

```bash
docker compose up --build          # Build and start
docker compose up -d --build       # Start in background
docker compose logs -f             # View logs
docker compose down                # Stop services
docker compose down -v             # Stop and remove volumes
docker compose build backend       # Rebuild specific service
```

## Project Structure

```
pageturner/
├── backend/
│   ├── internal/
│   │   ├── api/            # HTTP handlers, middleware, routes
│   │   ├── db/             # sqlc queries, schema, migrations
│   │   ├── repository/     # Data access layer
│   │   └── types/          # Shared domain types
│   ├── migrations/         # PostgreSQL migrations (goose)
│   ├── scripts/            # OpenLibrary data importer
│   ├── sqlc.yaml
│   └── Dockerfile
├── frontend/
│   └── web/
│       ├── src/routes/     # SvelteKit routes
│       └── Dockerfile
├── bruno/                  # API test collection
└── docker-compose.yml
```

## License

MIT