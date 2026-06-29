# VendoReport — Go API Server

A REST API server for managing Juanfi vending machines. Built with **Go**, **Gin**, and **GORM (SQLite)**.

---

## Table of Contents

- [System Requirements](#system-requirements)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Server](#running-the-server)
- [CLI Commands](#cli-commands)
- [API Reference](#api-reference)
- [Authentication](#authentication)
- [Cron Jobs](#cron-jobs)
- [Running Tests](#running-tests)
- [Production Deployment](#production-deployment)
- [CI/CD](#cicd)

---

## System Requirements

| Requirement | Minimum version |
|---|---|
| Go | 1.22+ |
| SQLite | 3.x (via CGO — no separate install needed) |
| GCC / Clang | Required by the SQLite CGO driver (`mattn/go-sqlite3`) |
| OS | Linux, macOS, Windows (WSL recommended on Windows) |

**macOS:**
```bash
xcode-select --install   # installs Clang
brew install go          # or download from https://go.dev/dl
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install golang gcc
```

**Verify:**
```bash
go version   # go1.22.0 or later
gcc --version
```

---

## Project Structure

```
server/
├── cmd/
│   └── main.go                  # Entry point — server + CLI dispatcher
├── app/
│   └── app.go                   # Wires all dependencies into a Gin router
├── commands/                    # Cobra CLI sub-commands
│   ├── root.go
│   ├── user_add.go
│   ├── vendo_add.go
│   ├── vendo_logger.go
│   └── vendo_status_log.go
├── internal/
│   ├── config/                  # Environment variable loading
│   ├── controllers/             # HTTP request/response handlers
│   ├── database/                # GORM connection setup
│   ├── middleware/               # JWT auth middleware
│   ├── models/                  # GORM model definitions
│   ├── repository/              # Database query layer
│   ├── scheduler/               # Cron job definitions
│   ├── services/                # Juanfi external API client + logger
│   └── websocket/               # WebSocket hub and handler
├── tests/
│   └── api_test.go              # Feature tests (all endpoints)
├── .env.example                 # Environment variable template
└── go.mod
```

---

## Installation

### 1. Clone the repository

```bash
git clone <repository-url>
cd VendoReport/server
```

### 2. Install Go dependencies

```bash
go mod tidy
```

### 3. Set up environment variables

```bash
cp .env.example .env
```

Edit `.env` with your values (see [Configuration](#configuration) below).

### 4. Build the binary (optional)

```bash
go build -o vendoreport ./cmd
```

Or run directly without building:

```bash
go run cmd/main.go
```

---

## Configuration

All configuration is loaded from `server/.env`. Copy `.env.example` to get started.

| Variable | Required | Default | Description |
|---|---|---|---|
| `DB_DRIVER` | No | `sqlite` | Database driver: `sqlite` or `mysql` |
| `DB_DSN` | Yes | — | SQLite file path **or** MySQL connection string (see below) |
| `APP_ENV` | No | `development` | `development` enables request logging; `production` disables it |
| `APP_PORT` | No | `8000` | HTTP server port |
| `JWT_SECRET` | Yes | insecure default | HS256 signing secret for JWT tokens |
| `CORS_ORIGINS` | No | — | Comma-separated list of allowed CORS origins |
| `TRUSTED_PROXIES` | No | — | Comma-separated IPs/CIDRs of upstream proxies (e.g. HAProxy). Enables correct `X-Forwarded-For` parsing. Leave empty when running without a reverse proxy |

**Generating a secure `JWT_SECRET`:**
```bash
openssl rand -hex 32
```

**Example `.env`:**
```
DB_DSN=../app.db
APP_ENV=production
APP_PORT=8000
JWT_SECRET=a3f8d2c1e9b74f6a1d2e3c4b5a6f7e8d9c0b1a2e3f4d5c6b7a8e9f0a1b2c3d4
CORS_ORIGINS=https://yourdomain.com
```

> The `DB_DSN` path is relative to the `server/` directory. `../app.db` points to the SQLite database at the project root, which is the same database used by the Python app.

---

## Running the Server

From the `server/` directory:

```bash
# Development (with request logging)
go run cmd/main.go

# Or run the compiled binary
./vendoreport
```

The server starts on `http://localhost:8000` by default.

**Startup output:**
```
2024/01/01 00:00:00 database: connected to ../app.db
2024/01/01 00:00:00 scheduler: started (notifications @1m, logs @5m, status @5m)
[GIN-debug] GET    /                         ...
[GIN-debug] POST   /token                    ...
...
2024/01/01 00:00:00 server: listening on :8000
```

---

## CLI Commands

Administrative commands are available as sub-commands. Run them from the `server/` directory.

### Add a user

```bash
go run cmd/main.go user-add
```
```
Enter username: admin
Enter password: secret
User "admin" was added.
```

### Register a vendo machine

```bash
go run cmd/main.go vendo-add
```
```
Enter vendo name: Mercy
Enter API URL (e.g. http://192.168.42.10:8081): http://192.168.42.10:8081
Enter API key: your-api-key
Vendo "Mercy" added (id=1).
```

### Manually pull logs from all active vendos

```bash
go run cmd/main.go vendo-logger
```

### Manually snapshot vendo status

```bash
go run cmd/main.go vendo-status-log
```

> These same operations run automatically via cron when the server is running. Use the CLI commands for one-off manual triggers or initial setup.

---

## API Reference

All protected endpoints require a `Bearer` token in the `Authorization` header.

### Authentication

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/` | No | Health check |
| `POST` | `/token` | No | Login and receive JWT |

### Users

| Method | Path | Auth | Permission | Description |
|---|---|---|---|---|
| `GET` | `/users/me` | Yes | any | Get currently authenticated user |
| `PUT` | `/users/me/password` | Yes | any | Change current user's password |
| `GET` | `/users` | Yes | `users` | List users (supports `?q=&sort_by=&sort_dir=`) |
| `GET` | `/users/:id` | Yes | `users` | Get user by ID |
| `POST` | `/users` | Yes | `users` | Create a user (`{username, password, role_id?, vendo_ids?, is_active?}`) |
| `PUT` | `/users/:id` | Yes | `users` | Update a user (all fields optional; cannot change own password here) |
| `DELETE` | `/users/:id` | Yes | `users` | Delete a user (self-delete returns 422) |

### Roles

| Method | Path | Auth | Permission | Description |
|---|---|---|---|---|
| `GET` | `/roles` | Yes | `users` | List all roles |
| `GET` | `/roles/:id` | Yes | `users` | Get role by ID |
| `POST` | `/roles` | Yes | `users` | Create a role (`{name, permissions: string[]}`) |
| `PUT` | `/roles/:id` | Yes | `users` | Update a role |
| `DELETE` | `/roles/:id` | Yes | `users` | Delete a role (returns 422 if assigned to any user) |

### Vendo Machines

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/vendo-machines` | Yes | List all vendos (supports `?q=name&is_active=true`) |
| `GET` | `/vendo-machines/:id` | Yes | Get a single vendo with latest status |
| `POST` | `/vendo-machines` | Yes | Register a new vendo machine |
| `DELETE` | `/vendo-machines/:id` | Yes | Delete a vendo machine |
| `GET` | `/vendo-machines/:id/status` | Yes | Fetch live status from the machine |
| `POST` | `/vendo-machines/:id/set-status` | Yes | Activate or deactivate a vendo |
| `POST` | `/vendo-machines/:id/withdraw-current-sales` | Yes | Reset current sales and record withdrawal |

### Sales & Logs

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/sales` | Yes | Paginated sales (`?q=&date=&vendo_id=&page=&size=`) |
| `GET` | `/daily-sales` | Yes | Aggregated daily sales (`?from_date=&to_date=`) |
| `GET` | `/monthly-sales` | Yes | Aggregated monthly sales (`?from_date=&to_date=`) |
| `GET` | `/logs` | Yes | Paginated system logs (`?q=&date=&vendo_id=&page=&size=`) |
| `POST` | `/log/refresh` | Yes | Manually trigger log pull for all active vendos |

### Status History & Withdrawals

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/vendo-status-history` | Yes | Hourly status snapshots (`?vendo_id=&from_date=&to_date=&active_only=`) |
| `GET` | `/withdrawals` | Yes | List all withdrawal records |

### WebSocket

| Path | Description |
|---|---|
| `GET /ws` | WebSocket connection for real-time notifications |

---

## Authentication

The API uses **JWT Bearer tokens** (HS256).

### Login

```bash
curl -X POST http://localhost:8000/token \
  -d "username=admin&password=yourpassword"
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiry": 1700000000,
  "token_type": "bearer",
  "user": { "id": 1, "username": "admin", "is_active": true }
}
```

### Using the token

```bash
curl -H "Authorization: Bearer <access_token>" \
  http://localhost:8000/vendo-machines
```

- Tokens are valid for **1 hour**
- A tampered or expired token returns `401 Unauthorized`

---

## Cron Jobs

The scheduler runs automatically inside the server process — no external cron setup required.

| Schedule | Job | Description |
|---|---|---|
| Every 1 minute | Notifications | Broadcasts unread notifications to WebSocket clients |
| Every 5 minutes | Log refresh | Fetches system logs and sales from all active vendos via Juanfi API |
| Every 5 minutes | Status snapshot | Records current uptime, sales, heap, signal strength per vendo |

The scheduler starts with the server and stops when the process exits. If you need to trigger these manually, use the [CLI commands](#cli-commands).

---

## Running Tests

Tests use an in-memory SQLite database — no setup, no external dependencies.

```bash
cd server
go test ./tests/... -v
```

```
=== RUN   TestHealthCheck
--- PASS: TestHealthCheck (0.00s)
=== RUN   TestLogin_Success
--- PASS: TestLogin_Success (0.00s)
...
PASS
ok  github.com/jariesdev/vendoreport/tests  20.6s
```

**57 tests** covering all endpoints, authentication, pagination, filters, permission enforcement, and error cases.

---

## Production Deployment

### 1. Build the binary

```bash
cd server
go build -o vendoreport ./cmd
```

### 2. Set up environment

```bash
cp .env.example .env
# Edit .env:
#   APP_ENV=production
#   JWT_SECRET=<openssl rand -hex 32>
#   DB_DSN=../app.db
```

### 3. Run with systemd

Create `/etc/systemd/system/vendoreportapi.service`:

```ini
[Unit]
Description=VendoReport Go API Server
After=network.target

[Service]
User=www-data
WorkingDirectory=/app/path
ExecStart=/app/path/go_app_be
EnvironmentFile=/app/path/.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable vendoreportapi
sudo systemctl start vendoreportapi
sudo systemctl status vendoreportapi
```

### 4. Run with pm2

```bash
pm2 start ./vendoreport --name vendoreport
pm2 save
pm2 startup
```

### Notes

- Set `APP_ENV=production` to disable request logging overhead
- The `JWT_SECRET` must be kept consistent across restarts — changing it invalidates all existing tokens
- The SQLite `app.db` file must be readable/writable by the process user
- The server binds to `0.0.0.0:<APP_PORT>` — put Nginx or a reverse proxy in front for HTTPS

---

## CI/CD

The workflow at `.github/workflows/server.yml` automates testing and deployment.

| Event | Jobs triggered |
|---|---|
| Push to any branch | `test` only |
| Pull request | `test` only |
| Push to `release/*` branch | `test` → `build` → `deploy` |
| Tag pushed (e.g. `v1.0.17`) | `test` → `build` → `deploy` |

### GitHub Variables

Set these under **Settings → Secrets and variables → Actions → Variables**:

| Variable | Example | Description |
|---|---|---|
| `PRODUCTION_SERVER_IP` | `192.46.225.21` | Server IP used for `ssh-keyscan` (known hosts) |
| `PRODUCTION_SERVER_URI` | `deploy@192.46.225.21` | SSH/SCP target (`user@host`) |
| `SERVER_ENV` | *(contents of `.env`)* | Written as `server/.env` on the production server |
| `SERVER_WORKING_DIR` | `/opt/vendoreport/server` | Directory where the binary runs |
| `SERVER_BUILD_DIR` | `/tmp/vendoreport` | Staging directory where the artifact is uploaded before extraction |
| `SERVER_SERVICE_NAME` | `vendoreport` | systemd service name (`systemctl restart <name>`) |

### GitHub Secrets

Set these under **Settings → Secrets and variables → Actions → Secrets**:

| Secret | Description |
|---|---|
| `SSH_PRIVATE_KEY` | Private SSH key with access to the production server |

### Deploying a release

```bash
git tag v1.0.17
git push origin v1.0.17
```

The workflow will test, build a stripped Linux binary, upload it to `SERVER_BUILD_DIR`, copy `.env` to `SERVER_WORKING_DIR`, and restart the systemd service.
