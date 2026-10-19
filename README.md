# IT-SUM Backend

IT-SUM is an educational platform designed for IT students. This repository contains the backend REST API built with Go, Gin, and PostgreSQL hosted on Neon.

## Architecture

The backend follows a clean, modular, and dependency-injected architecture designed for maintainability, security, and scalability:

```text
backend/
├── cmd/
│   ├── api/                # Main application entry point & HTTP server
│   │   └── main.go
│   └── migrate/            # Database migration CLI runner
│       ├── main.go
│       └── main_test.go
├── internal/
│   ├── config/             # Configuration loading (.env, env vars) & validation
│   │   ├── config.go
│   │   └── config_test.go
│   ├── database/           # PostgreSQL connection pooling with pgx/v5
│   │   ├── postgres.go
│   │   └── postgres_test.go
│   ├── http/
│   │   ├── handlers/       # HTTP route handlers (dependency-injected)
│   │   │   ├── health.go
│   │   │   └── health_test.go
│   │   └── routes/         # Versioned route registration (/api/v1)
│   │       ├── routes.go
│   │       └── routes_test.go
│   ├── middleware/         # HTTP middleware (CORS, logging, recovery)
│   │   ├── cors.go
│   │   └── cors_test.go
│   └── modules/            # Future business & domain modules
│       └── .gitkeep
├── migrations/             # SQL migration files
│   └── .gitkeep
├── .env.example            # Environment variables template
├── .gitignore              # Git ignore rules
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── README.md               # Project documentation
```

### Key Architectural Decisions

- **Dependency Injection**: Handlers declare interfaces (`DatabasePinger`) rather than relying on global state. This makes handlers modular and testable without requiring a live database.
- **Resource Pooling for Neon**: Neon PostgreSQL is serverless and scales computes down when inactive. Connection pooling via `pgxpool` uses conservative settings (`MaxConns: 5`, `MinConns: 0`, `MaxConnIdleTime: 5m`) to avoid connection exhaustion and accommodate compute suspension.
- **Sanitized Errors**: Database connection strings, passwords, and internal driver details are stripped from all logs and HTTP responses to prevent secret leakage.
- **Graceful Shutdown**: The server listens for `SIGINT` / `SIGTERM` signals, closes active HTTP connections within a 10-second timeout, and releases the database pool cleanly.

---

## Prerequisites

- **Go**: version 1.22+ (tested with Go 1.27)
- **PostgreSQL Database**: a PostgreSQL database or free-tier instance on [Neon](https://neon.tech)
- **Git**

---

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `PORT` | No | `8080` | Port for the HTTP server to listen on |
| `DATABASE_URL` | **Yes** | — | PostgreSQL connection string (`sslmode=require` for Neon) |
| `ALLOWED_ORIGINS`| No | `http://localhost:3000` | Comma-separated list of allowed frontend origins |
| `APP_ENV` | No | `development` | Application environment (`development` or `production`) |

### Example `.env`

Copy `.env.example` to `.env` for local development:

```bash
cp .env.example .env
```

Ensure `.env` contains:

```env
PORT=8080
APP_ENV=development
DATABASE_URL=postgres://neondb_owner:your_password_here@ep-example-pooler.us-east-2.aws.neon.tech/neondb?sslmode=require
ALLOWED_ORIGINS=http://localhost:3000
```

> **Security Warning**: Never commit `.env` or share credentials in logs, source code, or pull requests. `.env` is excluded by `.gitignore`.

---

## Local Setup & Running

1. **Clone the repository**:

   ```bash
   git clone https://github.com/TheSeifDev/it-sum.git
   cd it-sum/backend
   ```

2. **Install dependencies**:

   ```bash
   go mod download
   ```

3. **Configure environment**:

   ```bash
   cp .env.example .env
   # Edit .env with your PostgreSQL/Neon credentials
   ```

4. **Run the API server**:

   ```bash
   go run ./cmd/api
   ```

   The server will start on `http://localhost:8080`.

---

## Health Endpoints

The API provides two distinct health checks to separate application liveness from database readiness:

### 1. General Application Health (Liveness)

- **Endpoint**: `GET /api/v1/health`
- **Purpose**: Verifies that the HTTP service is alive and receiving traffic.
- **Response**: `200 OK`
  ```json
  {
    "service": "it-sum-api",
    "status": "ok"
  }
  ```

### 2. Database Connectivity Health (Readiness)

- **Endpoint**: `GET /api/v1/health/db`
- **Purpose**: Performs a bounded (3-second timeout) ping against PostgreSQL to confirm database readiness.
- **Success Response**: `200 OK`
  ```json
  {
    "database": "connected",
    "status": "ok"
  }
  ```
- **Failure Response**: `503 Service Unavailable`
  ```json
  {
    "database": "unavailable",
    "status": "error"
  }
  ```

### Testing with cURL / PowerShell

```bash
# General health
curl -i http://localhost:8080/api/v1/health

# Database health
curl -i http://localhost:8080/api/v1/health/db
```

In PowerShell:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/api/v1/health
Invoke-RestMethod -Uri http://localhost:8080/api/v1/health/db
```

---

## Database Migrations

Database migrations are managed using `golang-migrate` and are kept strictly decoupled from standard application startup.

### CLI Runner (`cmd/migrate`)

The project includes a built-in CLI tool to execute versioned migrations against your database:

```bash
# Check current migration version
go run ./cmd/migrate version

# Apply all pending migrations
go run ./cmd/migrate up

# Roll back the most recent migration step
go run ./cmd/migrate down

# Roll back all migrations
go run ./cmd/migrate down all

# Apply or roll back a specific number of steps (e.g., +1 or -1)
go run ./cmd/migrate step 1
go run ./cmd/migrate step -1
```

### Creating New Migrations

To add a new migration, place paired `.up.sql` and `.down.sql` files in the `migrations/` directory following the naming format:

```text
migrations/
├── 000001_create_example_table.up.sql
└── 000001_create_example_table.down.sql
```

Example migration syntax:

```sql
-- 000001_create_example_table.up.sql
CREATE TABLE IF NOT EXISTS example (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 000001_create_example_table.down.sql
DROP TABLE IF EXISTS example;
```

---

## Neon PostgreSQL Configuration

When connecting to Neon PostgreSQL:

1. **SSL Mode**: Always append `?sslmode=require` to your connection string.
2. **Connection Pooling**: Use Neon's pooled endpoint (`-pooler`) for web API workloads to benefit from PgBouncer transaction pooling.
3. **Conservative Pool Limits**:
   - `MaxConns: 5` ensures multiple API replicas do not exceed compute connection limits.
   - `MinConns: 0` allows idle connections to close so that Neon's serverless compute can enter suspend state when idle, conserving free-tier quota.
   - `MaxConnIdleTime: 5m` cleans up stale connections periodically.

---

## Deployment to Render (Free Tier)

To deploy the backend to Render:

1. **Create a Web Service** on Render connected to this repository.
2. Set the service properties:
   - **Root Directory**: `backend`
   - **Environment**: `Go`
   - **Build Command**: `go build -o bin/api ./cmd/api`
   - **Start Command**: `./bin/api`
3. **Environment Variables**:
   In the Render dashboard under Environment:
   - `PORT`: `10000` (or leave empty; Render injects `PORT`)
   - `APP_ENV`: `production`
   - `DATABASE_URL`: Your Neon PostgreSQL connection string (configured with `sslmode=require`)
   - `ALLOWED_ORIGINS`: Your frontend URL (e.g., `https://it-sum.vercel.app`)
4. **Health Check Path**:
   Set the Render health check path to `/api/v1/health`.

---

## Testing & Quality Assurance

Run the test suite and validation checks:

```bash
# Format code
gofmt -w .

# Tidy dependencies
go mod tidy

# Static analysis
go vet ./...

# Unit tests
go test -v ./...

# Compile all packages
go build ./...
```
