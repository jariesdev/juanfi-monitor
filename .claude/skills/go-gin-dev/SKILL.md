---
name: go-gin-dev
description: Use when writing, editing, or reviewing Go code under server/ — Gin handlers, services, repositories, GORM models, or the Juanfi API client. Covers layering conventions, run/test commands, and server-specific gotchas (SQLite migration, JWT, CGO).
---

# Go (Gin) server development

## Layer structure
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

## Conventions
* Write idiomatic Go and follow Go best practices.
* Use Gin only for HTTP routing, middleware, request binding, and response handling.
* Keep business logic in the service (use-case) layer.
* Keep Gin handlers/controllers thin: bind request → validate → call service → return response.
* Keep repositories responsible only for database access; never place business logic there.
* Use constructor dependency injection for all dependencies.
* Define interfaces only when necessary and where they are consumed.
* Always pass `context.Context` as the first parameter in service and repository methods (use `c.Request.Context()` from Gin).
* Return wrapped errors (`fmt.Errorf("...: %w", err)`) and let handlers map them to HTTP responses.
* Keep functions small, focused, and single-purpose.
* Prefer composition over unnecessary abstractions.
* Keep database transactions in the service layer.
* Use DTOs for request/response models and separate them from domain models.
* Validate request payloads in Gin handlers using struct tags; enforce business rules in services.
* Use table-driven tests and mock repositories/external services.
* Prefer the Go standard library unless a third-party package provides significant value.
* When formatting or parsing dates/times, use the named layout constants from the standard library `time` package (`time.DateOnly`, `time.DateTime`, `time.TimeOnly`, `time.RFC3339`, etc. — see `/usr/local/go/src/time/format.go`) instead of hardcoding the equivalent layout string. E.g. `time.ParseInLocation(time.DateOnly, dateStr, loc)` rather than `time.ParseInLocation("2006-01-02", dateStr, loc)`. Only write out a literal reference-time layout when no constant matches the format you need.

## Run & test
```bash
cd server
go run cmd/main.go          # start on :8000
go test ./tests/... -v      # 43 integration tests (in-memory SQLite)
go build -o vendoreport ./cmd
```

## Gotchas
See [[../../rules/server-gotchas.md]] for SQLite migration behavior, the GORM association crash, JWT secret requirement, and CGO build requirement.

## Feature-specific references
- Juanfi device API client — [[../../rules/juanfi-device-api.md]]
- Vendo Rates feature design — [[../../rules/vendo-rates-feature.md]]

## Environment variables (`server/.env`)
| Variable | Default | Notes |
|---|---|---|
| `DB_DRIVER` | `sqlite` | `sqlite` or `mysql` |
| `DB_DSN` | — | File path or MySQL DSN |
| `APP_ENV` | `development` | `production` disables request logging |
| `APP_PORT` | `8000` | |
| `JWT_SECRET` | — | Required; fatal in production if placeholder |
| `CORS_ORIGINS` | — | Comma-separated |
