# CI/CD & release conventions

## Workflows
- `.github/workflows/server.yml` — Go CI: test on push/PR, build + deploy on tag push or release branch push
- `.github/workflows/node.js.yml` — Node.js CI: build + deploy client-pwa on tag push or release branch push
- `.github/workflows/release.yml` — **Semantic Release**: runs on push to `main`, analyzes conventional commits, creates git tag + GitHub release

## Release flow
1. Feature branches → `release/x.y.z` → merge PR to `main`
2. On push to `main`, `.github/workflows/release.yml` runs `semantic-release`:
   - Analyzes commits since last tag using conventional commits
   - Determines next version (major/minor/patch)
   - Updates `CHANGELOG.md` via `@semantic-release/changelog`
   - Bumps version in `client-pwa/package.json` via `@semantic-release/npm`
   - Creates git tag + GitHub release via `@semantic-release/github`
   - Commits updated files via `@semantic-release/git`
3. The new git tag triggers `.github/workflows/server.yml` and `.node.js.yml` to build and deploy

## Required GitHub secrets
| Secret | Used by |
|---|---|
| `SSH_PRIVATE_KEY` | deploy jobs in server.yml & node.js.yml |
| `GH_TOKEN` | release.yml — must be a **PAT** (not GITHUB_TOKEN) so tag push triggers deploy workflows |

## Required GitHub variables
`PRODUCTION_SERVER_IP`, `PRODUCTION_SERVER_URI`, `SERVER_ENV`, `SERVER_WORKING_DIR`, `SERVER_BUILD_DIR`, `SERVER_SERVICE_NAME`, `PRODUCTION_WORKING_DIR`, `PRODUCTION_PM2_ID`

## Branch & release conventions
- Feature branches → `release/x.y.z` → `main`
- Tags trigger production deployment of the Go server binary

## Commit conventions
- Use concise, imperative tense in commit messages.
- Commits should be atomic and self-contained.
- Never include a `Co-Authored-By` trailer in commit messages.
- Use [Conventional Commits](https://www.conventionalcommits.org/) format (`feat:`, `fix:`, `chore:`, `docs:`, etc.) for automatic versioning via semantic-release.
