# Subscription Service

RESTful HTTP service for managing user subscriptions and aggregating their costs.

## Tech Stack

- **Language:** Go 1.26.3
- **Framework:** Gin
- **Database:** PostgreSQL 16 (via sqlx + lib/pq)
- **Swagger:** swaggo/gin-swagger
- **Auth/Build:** Docker Compose

## Architecture

```
handler -> service -> store -> PostgreSQL
   |         |         |
  HTTP    business    data
  format   logic      access
```

- **handler/** — request parsing, validation, HTTP responses
- **service/** — business rules, date conversion, delegation
- **store/** — SQL queries via sqlx
- **model/** — domain types (Subscription, ParseMonth)

## Quick Start

```bash
make up
```

Service will be available at `http://localhost:8080`.

## API

| Method | Path | Description |
|--------|------|-------------|
| POST   | /api/v1/subscriptions | Create subscription |
| GET    | /api/v1/subscriptions?page=1&limit=10 | List (paginated) |
| GET    | /api/v1/subscriptions/:id | Get by ID |
| PUT    | /api/v1/subscriptions/:id | Update |
| DELETE | /api/v1/subscriptions/:id | Delete |
| GET    | /api/v1/subscriptions/user/:user_id | Get by user |
| GET    | /api/v1/subscriptions/aggregate?start_date=MM-YYYY&end_date=MM-YYYY | Aggregate cost |

Swagger UI at `/swagger/index.html`.

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_HOST | localhost | Server host |
| SERVER_PORT | 8080 | Server port |
| DATABASE_HOST | localhost | PostgreSQL host |
| DATABASE_PORT | 5432 | PostgreSQL port |
| DATABASE_USER | — | PostgreSQL user |
| DATABASE_PASSWORD | — | PostgreSQL password |
| DATABASE_NAME | — | Database name |
| DATABASE_SSL_MODE | disable | SSL mode |

## Tests

```bash
make test
```
