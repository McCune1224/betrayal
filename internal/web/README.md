# Web Admin Panel

Echo v4 + templ + HTMX + Tailwind v4 admin panel for the Betrayal Bot. This file supersedes the root-level `WEB_ADMIN_PLAN.md` (folded in 2026-08); unimplemented pieces of that plan are listed under [Roadmap](#roadmap).

## Overview

Operational control + visibility for the game: session-auth'd dashboard, player/vote/role pages, command-audit viewer, and a Railway redeploy button. Runs inside the bot process (started from `cmd/betrayal-bot/main.go`); `make run-web` (`DISABLE_DISCORD=true`) boots **only** the web server for fast UI iteration with no Discord connection.

## Architecture

```
main.go (cmd/betrayal-bot)
 ├── Discord Bot (ken/discordgo)          Web Server (Echo)
 │        │                                    │
 │        └────────── Shared pgx Pool ─────────┘
 │        └────────── Shared discord.Session ───┘   (health handler pings Discord)
```

- `internal/web/server.go` — Echo setup, middleware, route table, graceful start/shutdown. `Server` holds the pgx pool, optional `*discordgo.Session`, session cookie store, and the Railway client.
- `internal/web/handlers/` — `auth.go` (login/logout), `dashboard.go`, `health.go`, `players.go`, `votes.go`, `roles.go`, `admin.go` (redeploy + audit), `format_args.go` (audit argument rendering).
- `internal/web/middleware/auth.go` — session-required guard applied to the protected route group.
- `internal/web/railway/client.go` — Railway GraphQL client used by the redeploy button.
- `internal/web/templates/` — templ layouts/pages/partials. Generated `_templ.go` files are **committed**, so `go build` works offline.
- `web/static/` — Tailwind source (`css/input.css`, v4 `@import "tailwindcss"`, "Dusty Western" theme in `@theme`) + generated `css/output.css` (committed) + vendored `htmx.min.js`.

## Routes

Public: `GET /login`, `POST /login`, `GET /health`. Everything else requires a session cookie.

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | `/logout` | auth | Clear session |
| GET | `/` | dashboard | Dashboard |
| GET | `/health/status` | health | HTMX health partial |
| GET | `/healthcheck` | readiness | Casual-user game readiness check (admin channels, funnels, players, cycle, lifeboard) |
| GET | `/players` | players | Player list page |
| GET | `/players/table` | players | HTMX player-table partial (search/filter) |
| GET | `/players/:id` | players | Player detail |
| GET | `/players/:id/edit` | players | Administrative player/inventory editor |
| POST | `/players/:id/edit` · `/players/:id/state` | players | Update stats and player state |
| POST | `/players/:id/{items,abilities,statuses,perks,immunities,notes}/*` | players | Inventory/immunity/note mutations, including shop purchase |
| GET | `/setup` | setup | Game role-pool generator |
| POST | `/setup/generate` | setup | Generate deceptionist options + random pool |
| GET | `/channels` | channels | Channel validation + configuration page |
| POST | `/channels/update` · `/channels/admin/delete` | channels | Configure vote/action/lifeboard/admin/log channels |
| GET | `/votes` | votes | Votes page |
| GET | `/votes/tally` | votes | HTMX vote tally partial |
| GET | `/roles` | roles | Roles list |
| GET | `/roles/search` | roles | HTMX role search partial |
| GET | `/roles/:id` | roles | Role detail |
| PUT | `/roles/:id` | roles | Update role |
| GET | `/roles/:id/abilities` · PUT `/roles/:id/abilities/:abilityId` · DELETE `/roles/:id/abilities/:abilityId` | roles | Ability sub-resources |
| GET | `/roles/:id/perks` · PUT `/roles/:id/perks/:perkId` · DELETE `/roles/:id/perks/:perkId` | roles | Perk sub-resources |
| POST | `/admin/redeploy` | admin | Trigger Railway redeploy |
| GET | `/admin/audit` | admin | Command audit log viewer |
| GET | `/sync` | sync | Spreadsheet sync page (sources, preview, history) |
| POST | `/sync/preview` | sync | Fetch + diff all enabled sources (read-only; rate-limited) |
| POST | `/sync/apply` | sync | Apply one source's diff (prod-guarded, rate-limited) |
| POST | `/sync/sources/:id` | sync | Edit source URL / enabled flag |
| GET | `/admin/migrations` | migrations | Embedded-migration status table + up/rollback controls |
| POST | `/admin/migrations/up` · `/admin/migrations/down` | migrations | Apply pending / roll back N (prod-guarded, rate-limited; rollback requires typing the migration name) |
| GET | `/admin/reset` | reset | Guided new-game reset preview with current row counts |
| POST | `/admin/reset` | reset | Atomically clear game state/catalog and reload all four CSV sources; requires explicit confirmation |

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `WEB_PORT` | Web server port (default 8080) |
| `ADMIN_PASSWORD` | Shared password for admin login |

| `WEB_ALLOW_PROD_MUTATIONS` | `true` lifts the hard-block on destructive panel actions (/sync/apply, migrations) against the prod pooler |
| `GOOD_ROLES_CSV` / `EVIL_ROLES_CSV` / `NEUTRAL_ROLES_CSV` / `ITEM_CSV` | Google Sheets CSV export URLs; seeded into `sync_source` at startup (editable in the /sync panel) |
| `RAILWAY_API_TOKEN` | Railway API token |
| `RAILWAY_BETRAYAL_PROJECT_ID` | Railway project ID |
| `RAILWAY_BETRAYAL_SERVICE_ID` | Railway service ID |
| `RAILWAY_BETRAYAL_ENVIRONMENT_ID` | Railway environment ID |
| `DISABLE_DISCORD` | `true` = web-only mode (no bot) |

## Working on Templates / CSS

- After editing `*.templ` or `web/static/css/input.css`: `make generate` (runs `templ generate` + `tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify`). Commit the generated `_templ.go` + `output.css`.
- Watch mode: `templ generate --watch` and `tailwindcss ... --watch` (see Makefile `templ-watch` / `tailwind-watch`).
- Tailwind is v4 CSS-first: theme tokens live in `@theme` in `input.css` — there is **no** `tailwind.config.js`.
- Theme direction: warm, non-corporate, **mobile-first** ("Dusty Western" palette). Preserve it.

## Security

Current state:
- Sessions: signed cookies via `gorilla/sessions` CookieStore; `HttpOnly` + `SameSite=Lax`, 7-day MaxAge.
- Session cookies are signed using a hash of `ADMIN_PASSWORD`; no second session credential is required.
- Startup applies all embedded migrations before the bot/web server begins
  serving requests. If migrations fail, startup fails closed instead of
  exposing routes backed by a partial schema.
TODO (required before adding any public-facing route):
1. CSRF middleware (Echo has built-in CSRF; must be wired to work with HTMX `HX-Request` headers).
2. Rate limiting on `/login` and `/admin/redeploy`.

## Roadmap

Not-yet-implemented items from the original `WEB_ADMIN_PLAN.md` / WT-6:
- Audit log pagination + filtering by command/user (plan Phase 5).
- Deployment status indicator next to the redeploy button (plan Phase 4, optional).
- Game-state admin pages: cycle control, channel-config validation (missing/orphaned channels), player edit (coins/items/status), item/ability/status CRUD (only roles exist today).
- Handler tests via Echo `httptest` (no web handler tests yet).

## History

- `WEB_ADMIN_PLAN.md` (repo root) was absorbed here on 2026-08 as part of the docs consolidation (WT-7). The plan's project-structure sketch is now realized; remaining deltas are the Roadmap above.
