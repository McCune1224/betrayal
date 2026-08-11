# Web Admin Panel

SvelteKit static web application served by the Go/Echo process. SvelteKit owns the complete UI; Go owns JSON APIs, authentication, Discord, PostgreSQL, migrations, sync/reset, and deployment operations.

## Architecture

- `frontend/` — full SvelteKit application, built with `adapter-static`.
- `internal/web/ui/dist/` — generated static output embedded into the Go binary.
- `internal/web/server.go` — Echo server, JSON API route table, browser auth boundary, and static shell fallback.
- `internal/web/api/` — authenticated JSON API handlers and explicit DTOs.
- `internal/web/middleware/auth.go` — browser session guard for the SvelteKit shell.
- `internal/web/railway/` — Railway GraphQL client used by the admin API.

Production has no SSR/Node service: the frontend is generated during the image build and served from the existing Go process.

## API and UI

All application data routes are authenticated JSON APIs under `/api/v1`:

- `/auth` — session, CSRF, login, logout.
- `/dashboard`, `/players` — dashboard, player list/detail/create/edit, inventory and note mutations.
- `/catalog` — roles, items, abilities, and statuses CRUD.
- `/ops` — cycle, channels, votes, readiness, and setup/role-pool generation.
- `/sync` — source listing/editing, preview, and apply.
- `/admin` — audit, migrations, reset, and Railway redeploy.

SvelteKit filesystem routes provide the corresponding pages, including `/login`, `/players`, `/players/new`, `/players/[id]`, `/players/[id]/edit`, catalog pages, operational pages, `/sync`, `/admin/audit`, `/admin/migrations`, `/admin/reset`, and `/admin/redeploy`.

Browser behavior is preserved at the boundary: unauthenticated `/` and client routes redirect to `/login`; unauthenticated `/api/v1/**` requests return canonical JSON `401` responses. Unknown API routes remain JSON `404`s.

## Build and development

- `make frontend-build` — install frontend dependencies and generate embedded static output.
- `make build` — generate the SvelteKit output and compile the Go binary.
- `make run-web` — run the Go web server without Discord.
- `npm --prefix frontend run test:unit` — frontend Vitest suite.
- `npm --prefix frontend run check` — Svelte diagnostics.
- `go test ./...` — Go tests against the isolated local database.

There are no templ, HTMX, or legacy Tailwind build steps. Do not add server-rendered HTML routes; add an API endpoint and SvelteKit route instead.

## Environment

| Variable | Purpose |
|---|---|
| `WEB_PORT` | Web server port, normally 8080 |
| `ADMIN_PASSWORD` | Shared admin password |
| `GOOD_ROLES_CSV`, `EVIL_ROLES_CSV`, `NEUTRAL_ROLES_CSV`, `ITEM_CSV` | Sync source URLs |
| `RAILWAY_API_TOKEN` and Railway IDs | Redeploy API configuration |
| `DISABLE_DISCORD` | `true` for web-only development |

## Security

- Signed `gorilla/sessions` cookies derived from `ADMIN_PASSWORD`.
- CSRF protection on unsafe requests, including JSON requests.
- Login, migration, sync, reset, and redeploy operations are rate limited where appropriate.
- Production sync/migration/reset behavior remains governed by the Go services and explicit confirmation checks.
