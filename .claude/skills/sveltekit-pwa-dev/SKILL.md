---
name: sveltekit-pwa-dev
description: Use when writing, editing, or reviewing code under client-pwa/ — SvelteKit 5 routes, components, or API calls. Covers the mandatory /x-api/ proxy rule, UIkit usage, auth flow, and run/test commands.
---

# client-pwa/ — SvelteKit 5 PWA development

## Conventions
- Use Component-Based Architecture (CBA) or Component-Driven Development (CDD).
- Use `src/lib/` for shared components, stores, and utilities.
- Use actions for form handling and validation.
- Use interfaces for data models and API responses.
- **UIkit is the primary UI component/styling library** — use `uk-*` classes (e.g. `uk-button`, `uk-input`, `uk-modal`, `uk-dropdown`) and the `uk-icon="icon: <name>"` directive for icons. It's installed as the `uikit` npm package (pinned to `3.18.3`), wired up in `src/routes/+layout.svelte` (`import 'uikit/dist/css/uikit.min.css'`, `import UIkit from 'uikit'`, `UIkit.use(icons)`). Check `node_modules/uikit/dist/js/uikit-icons.js` for available icon names before referencing one.

## API proxy rule — important
**All API calls must go through `/x-api/`**, never directly to the backend URL.

- The catch-all proxy at `src/routes/x-api/[...path]/+server.ts` reads the `auth_token` httpOnly cookie and injects `Authorization: Bearer <token>` on every request to the Go server.
- Do **not** import `baseApiUrl` from `$lib/env` in components — use `/x-api/<endpoint>` as the fetch URL.
- `VITE_API_URL` / `VITE_INTERNAL_API` are only used in server-side files (`+page.server.ts`, `+server.ts`, `vendo.remote.ts`).
- Exception: `src/routes/(guest)/login/+page.server.ts` calls `/token` directly server-side (needs to obtain the JWT before a cookie exists).

## Auth flow
1. Login page (`+page.server.ts`) POSTs credentials to Go `/token`, stores returned JWT as an httpOnly `auth_token` cookie.
2. Every subsequent request goes to `/x-api/*` → SvelteKit server reads the cookie → injects `Authorization: Bearer` → forwards to Go API.
3. The JWT never touches client-side JavaScript.

## Run & test
```bash
cd client-pwa
pnpm install --frozen-lockfile
pnpm run dev        # Vite dev server
pnpm run build      # production build → ./build/
pnpm run check      # svelte-check type check
pnpm run lint       # Prettier + ESLint
```

## Environment variables (`client-pwa/.env` or root `.env`)
| Variable | Used by | Notes |
|---|---|---|
| `VITE_API_URL` | server-side only | Go API URL (used in login + vendo.remote) |
| `VITE_INTERNAL_API` | server-side only | Go API URL for the x-api proxy |
| `VITE_WS_URL` | client-side | WebSocket URL |
