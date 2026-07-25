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

## Working in server/ or client-pwa/
Layering conventions, run/test commands, and gotchas live in dedicated skills — invoke them (or let them auto-trigger) when touching that area:
- `.claude/skills/go-gin-dev/SKILL.md` — Go/Gin server development
- `.claude/skills/sveltekit-pwa-dev/SKILL.md` — SvelteKit PWA frontend development

## Feature-specific references (`.claude/rules/`)
- [server-gotchas.md](.claude/rules/server-gotchas.md) — SQLite migration behavior, GORM association crash, JWT/CGO requirements
- [juanfi-device-api.md](.claude/rules/juanfi-device-api.md) — Juanfi device API client, rate-plan string format
- [vendo-rates-feature.md](.claude/rules/vendo-rates-feature.md) — `vendo_rates` table design, ACL, sync/import semantics
- [ci-cd-release.md](.claude/rules/ci-cd-release.md) — GitHub workflows, semantic-release flow, required secrets/vars, branch & commit conventions
- [documentation-standard.md](.claude/rules/documentation-standard.md) — doc structure/style standard for this repo
