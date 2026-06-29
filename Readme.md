
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

## Database

Schema is managed directly via SQL against `app.db` (SQLite). No migration tool is used — changes to the schema should be applied manually or via a SQL script.

For MySQL, set `DB_DRIVER=mysql` and `DB_DSN=<dsn>` in `server/.env`; GORM AutoMigrate runs automatically on startup.
