# VendoReport — Project Guide for Claude

## Repository layout

```
VendoReport/
├── server/          # Go REST API (Gin + GORM + SQLite/MySQL)
├── client-pwa/      # SvelteKit 5 PWA frontend
├── docker/          # Dockerfiles (node/)
├── docker-compose.yml
└── app.db           # SQLite database (shared by server)
```
## Permissions
The application is permission-based ACL. Roles are created dynamically via the UI or API — there are no hardcoded role names. Each role has a set of permissions, which grant access to specific features.

---

## server/ — Go API

### Go (Gin) Development Rules

* Write idiomatic Go and follow Go best practices.
* Use Gin only for HTTP routing, middleware, request binding, and response handling.
* Keep business logic in the service (use-case) layer.
* Keep Gin handlers/controllers thin: bind request → validate → call service → return response.
* Keep repositories responsible only for database access; never place business logic there.
* Use constructor dependency injection for all dependencies.
* Define interfaces only when necessary and where they are consumed.
* Always pass context.Context as the first parameter in service and repository methods (use c.Request.Context() from Gin).
* Return wrapped errors (fmt.Errorf("...: %w", err)) and let handlers map them to HTTP responses.
* Keep functions small, focused, and single-purpose.
* Prefer composition over unnecessary abstractions.
* Keep database transactions in the service layer.
* Use DTOs for request/response models and separate them from domain models.
* Validate request payloads in Gin handlers using struct tags; enforce business rules in services.
* Use table-driven tests and mock repositories/external services.
* Prefer the Go standard library unless a third-party package provides significant value.
* Generate production-ready, maintainable, and testable code by default.

### Run & test
```bash
cd server
go run cmd/main.go          # start on :8000
go test ./tests/... -v      # 43 integration tests (in-memory SQLite)
go build -o vendoreport ./cmd
```

### Key rules
- **SQLite schema** — `database.Connect()` takes a `migrate bool`. `config.ShouldMigrate()` now always returns `true` (the old Python/Alembic app is gone), but for SQLite this only ever *creates missing tables* (`sqliteCreateMissing()` in `internal/database/database.go`) — it never alters an existing table.
- **GORM association gotcha (SQLite)** — do **not** add a `gorm:"foreignKey:..."` association struct field (e.g. `Vendo *Vendo`) to a new model unless the code actually dereferences it. `AutoMigrate` on a model with such an association also walks into and tries to reconcile the *referenced* table's schema — and since the real `app.db`'s tables were originally created by the old Alembic app, they don't column-for-column match what GORM expects. SQLite's migrator then tries to rebuild the referenced table (rename → create temp → copy → drop) and the generated `INSERT...SELECT` can silently omit a NOT NULL column, crashing the server on startup with e.g. `NOT NULL constraint failed: vendos__temp.name`. Keep new models referencing existing tables to a plain `VendoID *uint` scalar column — no association struct field — unless you actually need to preload it.
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

### Juanfi device API (`internal/services/juanfi_api.go`)
All vendo machines run Juanfi firmware. Every call goes to `<api_url>/admin/<path>` with an `X-TOKEN: <api_key>` header and a `query=<unix-ms-timestamp>` param auto-appended (`buildURL`). Most endpoints are GET-only, even ones that mutate device state (e.g. `ResetCurrentSales` is a GET with `?type=coinCount`) — `api/saveRates` is the one exception, requiring a real POST.

Rate plans (`api/getRates` / `api/saveRates`) use a pipe/hash-delimited string format: entries separated by `|`, fields within an entry separated by `#` — `Name#Price#Minutes#ValidityMinutes#DataLimitMB#UserProfile` (`DataLimitMB` and `UserProfile` are optional/may be blank). `GetRates()` parses this into `[]Rate`; `SaveRates()` (`encodeRates`) serialises it back, POSTing `data=<encoded string>` as `application/x-www-form-urlencoded` with `rateType=1`. A blank `UserProfile` from the device means Mikrotik's `default` hotspot profile — `GetRates` normalizes it to the literal string `"default"` rather than `""`.

### Vendo Rates feature
`vendo_rates` table (`internal/models/vendo_rate.go`) stores per-vendo pricing tiers, imported from or pushed to the device via the Juanfi API above. Design:
- `vendo_id` is nullable — rows with `vendo_id = NULL` are the shared **default rate-plan template**, managed only by admins (`PermUsers`/`assignedVendoIDs(c) == nil`), separate from per-vendo rates which any user with access to that vendo can manage.
- **ACL** — all rate routes live under a `RequirePermission(models.PermRates)` (`"rates"`) group in `app.go`, so the `rates` permission is the entry ticket for the whole feature. On top of that, controllers enforce finer ACL: per-vendo ops also require `canAccessVendo`, and default-template/`set-as-default`/`apply-to-all` also require admin (`isAdmin`). Net effect: managing the default template needs **both** `rates` and `users` — the seed admin role (`AllPermissions()`) has both. Frontend mirrors this: per-vendo rates page (`/vendo/[id]/rates`) and the VendoTable "rates" row action gate on `rates`; the Settings → Default Rates page (`/settings/default-rates`) gates on `rates` **and** `users`.
- `SaveRates` (`POST /vendo-machines/:id/rates/sync`, the "Sync" button) uploads the vendo's stored rows to the device; `SetAsDefault` accepts a `mode` body of `copy` (append to the default template) or `replace` (default).
- All copy/import operations (`ImportFromMachine`, `SetAsDefault` replace mode, `ApplyToAll` in `vendo_rate_controller.go`) use **replace semantics** — they delete the target's existing rows before inserting the new set.
- `POST /vendo-rates/apply-to-all` requires an explicit, non-empty `vendo_ids` list — there's no "omit it to mean all" behavior; the frontend's vendo picker defaults to none selected with a "Select all" toggle.
- Frontend: per-vendo CRUD at `/vendo/[id]/rates`, global default-template admin page at `/settings/default-rates`.

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
- Use Component-Based Architecture (CBA) or Component-Driven Development (CDD)
- Use `src/lib/` for shared components, stores, and utilities
- Use actions for form handling and validation
- Use interfaces for data models and API responses
- **UIkit is the primary UI component/styling library** — use `uk-*` classes (e.g. `uk-button`, `uk-input`, `uk-modal`, `uk-dropdown`) and the `uk-icon="icon: <name>"` directive for icons. It's vendored locally, not an npm package — CSS at `src/lib/css/uikit*.css` and JS at `src/lib/js/uikit*.js`/`uikit-icons*.js`. The bundled icon set is smaller than UIkit's full set (e.g. there's `eye-slash` but no plain `eye`) — check `src/lib/js/uikit-icons.js` for available icon names before referencing one.

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

## Commit conventions
- Use concise, imperative tense in commit messages.
- Commits should be atomic and self-contained.
- Never include a `Co-Authored-By` trailer in commit messages.

## Documentation Standard

When generating or updating documentation, produce documentation that is optimized for both human developers and AI assistants.

### Principles

- Write for humans first; structure for AI retrieval.
- Be concise, explicit, and avoid unnecessary prose.
- Explain both **what** the component does and **why** it exists.
- State assumptions, constraints, and invariants explicitly.
- Use consistent terminology throughout the project.
- Prefer short, self-contained sections over long narratives.
- Use Markdown headings, bullet points, and tables where appropriate.
- Make every document understandable without requiring unrelated context.

### Standard Structure

Each document should include, when applicable:

1. Overview
    - Purpose
    - Responsibilities
    - Scope

2. Architecture
    - Position within the system
    - Upstream dependencies
    - Downstream consumers

3. Data Flow
    - Inputs
    - Processing
    - Outputs

4. Business Rules
    - Validation rules
    - Constraints
    - Assumptions
    - Invariants

5. Dependencies
    - Internal modules
    - External services
    - Database tables
    - Configuration

6. API / Interface
    - Public methods/endpoints
    - Parameters
    - Return values
    - Error cases

7. Security
    - Authentication
    - Authorization
    - Sensitive data handling

8. Performance
    - Caching
    - Concurrency
    - Scalability considerations

9. Failure Modes
    - Expected failures
    - Recovery strategy
    - Logging and monitoring

10. Related Components

11. Future Considerations
    - Known limitations
    - Technical debt
    - Planned improvements

### Documentation Rules

- Keep documents modular and focused on a single topic.
- Avoid ambiguous or inferred behavior—be explicit.
- Document architectural decisions and the rationale behind them.
- Cross-reference related documents instead of duplicating content.
- Prefer examples over lengthy explanations.
- Include Mermaid diagrams when they improve understanding.
- Treat documentation as part of the codebase and keep it synchronized with implementation.