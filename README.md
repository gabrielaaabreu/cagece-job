# Water consumption service (backend)

This workspace contains a simple Go backend service that stores users and their monthly water consumption records in PostgreSQL. The service is containerized and runnable with Docker Compose.

Quick start

1. Build and start services:

```bash
docker compose up --build
```

2. API endpoints (default localhost:3000):

- GET / -> service info
- POST /users -> create user
  - body: { "name": "...", "email": "..." }
- GET /users -> list users
- GET /users/{id} -> get user
- POST /users/{id}/consumptions -> add monthly consumption
  - body: { "year": 2025, "month": 10, "cubic_meters": 12.34 }
- GET /users/{id}/consumptions -> list for a user
- GET /consumptions -> list all (optional query params: user_id, year, month)

Examples

Create a user:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"name":"Alice","email":"alice@example.com"}' http://localhost:3000/users
```

Add consumption for user 1:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"year":2025,"month":10,"cubic_meters":15.2}' http://localhost:3000/users/1/consumptions
```

List consumptions:

```bash
curl http://localhost:3000/consumptions
```

Notes

- Database connection can be configured via `DATABASE_URL` or `PGHOST`/`PGUSER`/`PGPASSWORD`/`PGDATABASE` environment variables.
- The service will create the required tables on startup if they are missing.
