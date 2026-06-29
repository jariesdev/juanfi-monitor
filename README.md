
# VendoReport

## Dev

### API Server (Go)
- Copy `server/.env.example` to `server/.env` and fill in values
- `cd server && go run cmd/main.go` — starts on http://localhost:8000
- Requires `gcc` for CGO (`brew install gcc` on macOS)

### Frontend (SvelteKit)
- `cd client-pwa && pnpm install --frozen-lockfile`
- `pnpm run dev` — starts on http://localhost:5173

### Docker (both services)
- `docker compose up --build`

---

## Production

### API Server
- Build: `cd server && CGO_ENABLED=1 go build -o vendoreport ./cmd`
- Deploy the `vendoreport` binary and `server/.env` to the server
- Systemd service example:
  ```
  [Unit]
  Description=VendoReport API
  After=network.target

  [Service]
  User=www-data
  WorkingDirectory=/opt/vendoreport
  ExecStart=/opt/vendoreport/vendoreport
  EnvironmentFile=/opt/vendoreport/.env
  Restart=on-failure

  [Install]
  WantedBy=multi-user.target
  ```
- `systemctl enable --now vendoreport`

### Frontend
- Build: `cd client-pwa && pnpm run build` → outputs to `client-pwa/build/`
- Serve with Node/pm2:
  ```
  pm2 start client-pwa/build/index.js --name vendo-app
  ```

---

## Roles & Permissions

Access is controlled by roles assigned to users. Roles are created dynamically via the UI or API — there are no hardcoded role names.

### Permission list

| Permission | Grants access to |
|---|---|
| `dashboard` | Dashboard (default for all users) |
| `account` | Account / profile page (default for all users) |
| `vendos` | Vendo machine list, status, and active-users pages |
| `sales` | Sales reports and daily/monthly aggregates |
| `logs` | System logs |
| `withdrawals` | Withdrawal records |
| `users` | User management, role management (`/users` and `/roles` API endpoints; Users & Roles pages in the UI) |

**Default:** a user with no role assigned implicitly receives `dashboard` and `account`.

**Vendo scoping:** users without the `users` permission only see the vendo machines explicitly assigned to them.

### First-time setup

Create an admin role via the API after starting the server:

```bash
# 1. Create the admin role
curl -X POST http://localhost:8000/roles \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin","permissions":["dashboard","account","vendos","sales","logs","withdrawals","users"]}'

# 2. Assign the role to a user (use the role_id returned above)
curl -X PUT http://localhost:8000/users/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"role_id":1}'
```

---

## Database

Schema is managed directly via SQL against `app.db` (SQLite). No migration tool is used — changes to the schema should be applied manually or via a SQL script.

For MySQL, set `DB_DRIVER=mysql` and `DB_DSN=<dsn>` in `server/.env`; GORM AutoMigrate runs automatically on startup.
