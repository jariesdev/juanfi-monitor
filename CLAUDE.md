# VendoReport — Project Guide for Claude

## Repository layout

```
VendoReport/
├── server/          # Go REST API (Gin + GORM + SQLite/MySQL)
├── client-pwa/      # SvelteKit 5 PWA frontend
├── app.db           # SQLite database (shared by server)
└── migrations/      # Alembic migrations (schema source of truth for SQLite)
```

---

## server/ — Go API

### Run & test
```bash
cd server
go run cmd/main.go          # start on :8000
go test ./tests/... -v      # 43 integration tests (in-memory SQLite)
go build -o vendoreport ./cmd
```

### Key rules
- **SQLite schema is owned by Alembic** — never run `AutoMigrate` against the real `app.db`. `database.Connect()` takes a `migrate bool`; pass `false` for SQLite, `true` for MySQL and tests.
- **JWT secret** — must not be `change_me_in_production` when `APP_ENV=production`; the app will fatal on startup.
- **CGO required** — `mattn/go-sqlite3` needs gcc. Set `CGO_ENABLED=1` when building for deployment.
- All authenticated routes expect `Authorization: Bearer <JWT>` (HS256, 1-hour expiry).

### Layer structure
```
cmd/main.go          → wires config, DB, app, graceful shutdown
app/app.go           → single router wiring (used by server + tests)
internal/
  config/            → env var loading (Config struct)
  controllers/       → HTTP handlers
  repository/        → DB queries (interface + impl, SOLID DI)
  services/          → Juanfi external API client + logger
  scheduler/         → cron jobs (notifications @1m, logs/status @5m)
  middleware/        → JWT auth (reads Authorization header)
  websocket/         → notification broadcast hub
commands/            → Cobra CLI (user-add, vendo-add, vendo-logger, vendo-status-log)
tests/api_test.go    → feature tests, in-memory SQLite, seeds data
```

### Environment variables (`server/.env`)
| Variable | Default | Notes |
|---|---|---|
| `DB_DRIVER` | `sqlite` | `sqlite` or `mysql` |
| `DB_DSN` | — | File path or MySQL DSN |
| `APP_ENV` | `development` | `production` disables request logging |
| `APP_PORT` | `8000` | |
| `JWT_SECRET` | — | Required; fatal in production if placeholder |
| `CORS_ORIGINS` | — | Comma-separated |

---

## client-pwa/ — SvelteKit 5 PWA

### Run & test
```bash
cd client-pwa
pnpm install --frozen-lockfile
pnpm run dev        # Vite dev server
pnpm run build      # production build → ./build/
pnpm run check      # svelte-check type check
pnpm run lint       # Prettier + ESLint
```

### API proxy rule — important
**All API calls must go through `/x-api/`**, never directly to the backend URL.

- The catch-all proxy at `src/routes/x-api/[...path]/+server.ts` reads the `auth_token` httpOnly cookie and injects `Authorization: Bearer <token>` on every request to the Go server.
- Do **not** import `baseApiUrl` from `$lib/env` in components — use `/x-api/<endpoint>` as the fetch URL.
- `VITE_API_URL` / `VITE_INTERNAL_API` are only used in server-side files (`+page.server.ts`, `+server.ts`, `vendo.remote.ts`).
- Exception: `src/routes/(guest)/login/+page.server.ts` calls `/token` directly server-side (needs to obtain the JWT before a cookie exists).

### Auth flow
1. Login page (`+page.server.ts`) POSTs credentials to Go `/token`, stores returned JWT as an httpOnly `auth_token` cookie.
2. Every subsequent request goes to `/x-api/*` → SvelteKit server reads the cookie → injects `Authorization: Bearer` → forwards to Go API.
3. The JWT never touches client-side JavaScript.

### Environment variables (`client-pwa/.env` or root `.env`)
| Variable | Used by | Notes |
|---|---|---|
| `VITE_API_URL` | server-side only | Go API URL (used in login + vendo.remote) |
| `VITE_INTERNAL_API` | server-side only | Go API URL for the x-api proxy |
| `VITE_WS_URL` | client-side | WebSocket URL |

---

## CI/CD

Workflow: `.github/workflows/server.yml`
- **Push to any branch / PR** → runs Go tests only
- **Tag push** (e.g. `git tag v1.0.17 && git push --tags`) → test → build → deploy via SSH + systemd

Required GitHub variables: `PRODUCTION_SERVER_IP`, `PRODUCTION_SERVER_URI`, `SERVER_ENV`, `SERVER_WORKING_DIR`, `SERVER_BUILD_DIR`, `SERVER_SERVICE_NAME`
Required GitHub secret: `SSH_PRIVATE_KEY`

---

## Branch & release conventions
- Feature branches → `release/x.y.z` → `main`
- Tags trigger production deployment of the Go server binary
