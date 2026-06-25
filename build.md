# Docker Build Guide

This document describes how to build and run the bookTracker application using Docker and Docker Compose.

## Architecture

The application consists of three services:

- **PostgreSQL** (port 5432): Database for storing book data
- **Backend** (port 8080): Go API server
- **Frontend** (port 5173): SvelteKit web application

## Prerequisites

- Docker
- Docker Compose

## Quick Start

To build and start all services:

```bash
docker-compose up --build
```

Once running, access the application at: http://localhost:5173

## Services

### PostgreSQL
- Image: `postgres:16-alpine`
- Default credentials:
  - User: `booktracker`
  - Password: `booktracker`
  - Database: `booktracker`
- Data persisted in Docker volume `postgres_data`

### Backend
- Built from `backend/Dockerfile`
- Runs Go application on port 8080
- Connects to PostgreSQL via Docker networking
- Environment variables (can be overridden):
  - `DB_HOST`: postgres (Docker service name)
  - `DB_PORT`: 5432
  - `DB_USER`: booktracker
  - `DB_PASSWORD`: booktracker
  - `DB_NAME`: booktracker
  - `SSL_MODE`: disable

### Frontend
- Built from `frontend/web/Dockerfile`
- Runs SvelteKit application on port 5173
- Environment variables:
  - `PUBLIC_API_URL`: http://backend:8080 (internal Docker URL)
  - `ORIGIN`: http://localhost:5173
  - `DATABASE_URL`: /data/app.db (SQLite for auth)
  - `BETTER_AUTH_SECRET`: authentication secret

## Networking

All services communicate via the `booktracker-network` Docker bridge network:

- Frontend connects to backend using internal URL `http://backend:8080`
- Backend connects to PostgreSQL using service name `postgres`
- External ports are exposed for local access only

## Commands

### Build and start all services
```bash
docker-compose up --build
```

### Start in detached mode
```bash
docker-compose up -d --build
```

### View logs
```bash
docker-compose logs -f
```

### Stop all services
```bash
docker-compose down
```

### Stop and remove volumes (reset database)
```bash
docker-compose down -v
```

### Rebuild specific service
```bash
docker-compose build backend
docker-compose build frontend
```

### Run a specific service
```bash
docker-compose up postgres
docker-compose up backend
docker-compose up frontend
```

## Troubleshooting

### Backend cannot connect to database
- Ensure PostgreSQL container is running: `docker-compose ps`
- Check logs: `docker-compose logs postgres`
- Wait for health check to pass (may take a few seconds on first start)

### Frontend shows connection errors
- Verify backend is running: `docker-compose ps`
- Check frontend logs: `docker-compose logs frontend`
- Verify PUBLIC_API_URL is set correctly in environment

### Reset everything
```bash
docker-compose down -v
docker-compose up --build
```

### Access container shell for debugging
```bash
docker exec -it booktracker-backend sh
docker exec -it booktracker-frontend sh
docker exec -it booktracker-db psql -U booktracker -d booktracker
```