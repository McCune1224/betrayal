# Agent Guidelines for Betrayal Bot

## What this bot is

Discord game-management bot for "Betrayal" (battle-royale game). Go 1.23, discordgo + zekroTJA/ken (slash commands), pgx/v5 + sqlc (`internal/models/`), Echo + templ + HTMX + Tailwind v4 web admin panel, zerolog logging with DB audit trail. Hosted on **Railway** (prod). The legacy Fly workflow was removed (2026-08) — do not reintroduce it.

## Build & Run

**Prereqs (one-time per machine):** `templ` CLI (pin v0.3.960: `go install github.com/a-h/templ/cmd/templ@v0.3.960`), Tailwind v4.1.2 standalone binary (`~/bin/tailwindcss`), golang-migrate CLI with the postgres driver (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1`), and `$HOME/go/bin:$HOME/bin` on PATH.

- **Full bot** (Discord + web): `make run` — requires `.env` (see Worktrees & Env).
- **Web panel only**: `make run-web` (sets `DISABLE_DISCORD=true`) — fastest way to iterate on the admin UI, no Discord needed.
- **Build**: `make build` → templ generate + tailwind build + `go build` to `./bin/`.
- **Generate assets**: `make generate` (templ + tailwind) — required after editing `*.templ` or `web/static/css/input.css`.
- **Hot reload**: `air` is supported (`.air.toml` is gitignored); pair with `templ generate --watch` + `tailwindcss --watch` for template/CSS iteration.
- **Tests**: `go test ./...` — REQUIRES local Postgres (`make db-up` first, then `make mock-migrate-up`). Tests must never touch the production DB.
- **Migrations**: `make migrate-up/down` (prod via `DATABASE_POOLER_URL`), `make mock-migrate-up/down` (local via `MOCK_DATABASE`). Never run migrate-up against prod casually.
- **Local DB**: `make db-up` / `make db-down` (docker compose, postgres:16 on 5432). **Run `db-up` ONCE per machine — all worktrees share the same compose container** (stable name `betrayal-postgres`, so `docker exec betrayal-postgres ...` works from any worktree). Running `db-up` from a second worktree fails with a container-name conflict — that's expected. No Redis — Ken state is internal.

## Worktrees & Env (READ FIRST)

- `.env` is gitignored; each git worktree lands with no config. **Single source of truth lives outside the repo** at `~/.config/betrayal/env` (override with `BETRAYAL_ENV_FILE`).
- Every checkout's `.env` is a **symlink** to that file — godotenv autoload, Makefile greps, and tests all keep working unchanged.
- Tooling: `scripts/dev-env.sh` (also `make env-link` / `make worktree name=...`):
  - `link` — symlink this checkout's `.env` → canonical file (scaffolds from `.env.example` if missing).
  - `new-worktree NAME` — `git worktree add ../betrayal-NAME -b wt-NAME`, links env, boots local DB, runs mock migrations.
  - `doctor` — verifies env file, required keys, DB config.
- **Per-worktree overrides**: if one worktree needs different creds (e.g. a different bot token), replace its symlink with a real `.env` file. NEVER edit the shared canonical file for a one-off, and never `git add -f` any `.env`.
- Required keys: see `.env.example`. `ENVIRONMENT=local` in dev (console logs); `production` writes logs to the DB.
- If `dev-env.sh doctor` fails, fix the reported key before running anything.

## Repo Layout

- `cmd/betrayal-bot/main.go` — wiring: config from env, DB pool, Ken init, **command registration list**, web server. Add new commands HERE.
- `cmd/audit-analysis/`, `cmd/data-entry/` — standalone CLI tools (build with `go build ./cmd/{name}`; never commit binaries).
- `internal/commands/{name}/` — ken command packages (struct implements `ken.Command` + `Initialize(*pgxpool.Pool)`; keep the `var _ ken.SlashCommand = (*X)(nil)` assertion).
- `internal/services/` — reusable game logic (inventory exists). Keep new rules here so they're unit-testable WITHOUT Discord.
- `internal/models/` — sqlc-generated query code. Edit `internal/db/query/*.sql` and regenerate; do NOT hand-edit `*.sql.go`.
- `internal/db/migration/` — golang-migrate files. Name new ones `NNNN_name.up.sql` / `NNNN_name.down.sql` (legacy files 000016–000018 have inconsistent names; do not rename applied migrations).
- `internal/discord/` — embed/error/component/channel helpers. (`channels.go` was renamed from the legacy `chanenls.go` typo.)
- `internal/web/` — Echo server, handlers, middleware, Railway client, templ templates. Static assets in `web/static/` (Tailwind input/output.css, vendored `htmx.min.js`).
- `tests/` — testify suites. DB suites need local Postgres; logger suites are pure unit tests; web handler tests use Echo httptest.
- `scripts/dev-env.sh` — worktree/env tooling (see above).

## Command Inventory (registered in `main.go`)

| Command | Purpose | Admin? |
|---------|---------|--------|
| `/inv` | inventory mgmt: ability/item/coin/status/perk/alignment/role/immunity/luck/death/notes + create | most subcommands |
| `/roll` | rolls: item/ability rarity, player choice, luck, event rolls | hybrid |
| `/action` | submit game action (confessional) | player |
| `/view` | view role/ability/item/status details with buttons | player |
| `/buy` | purchase item for player | player |
| `/channel` | channel config: admin/vote/action/lifeboard/confessionals | admin |
| `/help` | help embeds (player + admin) | both |
| `/vote` | cast votes (funnel channel) | player |
| `/setup` | generate role list from CSV data entry | admin |
| `/echo` | ping/debug | admin |
| `/list` | list roles/items/statuses/etc | player |
| `/search` | fuzzy search abilities/items/statuses | player |
| `/healthcheck` | bot health | admin |
| `/cycle` | current/next/set phase + broadcast to confessionals/funnels/alliances | admin |
| `/tarot` | tarot draws (deterministic/per-user/guild-deck/random) | both |

**Admin roles** (`internal/discord/role.go`): Host, Co-Host, Bot Developer — check with `discord.IsAdminRole(ctx, discord.AdminRoles...)`, respond with `discord.NotAdminError(ctx)`.

**Auth pattern for channel commands**: `discord.IsAdminRole` gate + `util.ErrorContains/ErrorNotFound` helpers; log with `logger.Get().Error().Err(err).Msg(...)`; respond via `discord.SuccessfulMessage` / `discord.ErrorMessage` / `discord.AlexError`.

## Web Admin Panel

- Routes (`internal/web/server.go`): `/login`, `/` dashboard, `/health`, `/players` + `/players/:id`, `/votes`, `/roles` CRUD, `/admin/audit`, `/admin/redeploy` (Railway). Session-auth protected except `/login` + `/health`.
- Editing templates requires `make generate` (commits `_templ.go`); Tailwind source is v4 CSS (`@import "tailwindcss"`, theme tokens in `@theme` — "Dusty Western" palette).
- Theme: warm, non-corporate, **mobile-first** — preserve this.
- Security TODOs (before adding public-facing routes): CSRF middleware, login rate limiting, require `SESSION_SECRET` (no password fallback).

## Known Jank Register (fix under WT-5, don't perpetuate)

- `main.go:136` sets Intents from a permission constant (`PermissionAdministrator`) — verify/fix gateway intents (`IntentsAll` or explicit set).
- `roll.go:293` ability-roll path broken (FIXME).
- `view.go:213` categories FIXME; `help/player.go:49` button-builder FIXME ("What the actual hell").
- `inv/create.go` hardcoded game constants (coins 200 / items 4 / luck 0) + "unholy" switch chains.
- `main.go:316` hardcoded command-log channel ID (migration 000028 planned for configurability).
- `internal/services/inventory/inventory.go` `Jank()` is a documented hack — prefer `NewInventoryHandler`.
- Logger `Init` + Ken `Unregister` happen twice at startup (cleanup planned).
- `context.TODO()` in service writes (`item.go:15`, `status.go:15`) — thread real contexts.

## Deployment

- Prod = **Railway** (env-driven; in-app redeploy button via `internal/web/railway`). `.github/workflows/fly-deploy.yml` was deleted 2026-08.
- Dockerfile: single-stage today; builds templ + tailwind at image build; needs no `.env` at build (runtime env only).
- Never commit binaries (`betrayal-bot`, `bin/`) or `.env` files — `make clean` removes build artifacts.

## Testing Workflow (keep it fast)

1. `make db-up` (docker compose postgres) once per machine/worktree; `make mock-migrate-up`.
2. `go test ./...` — unit + DB suites against LOCAL `DATABASE_URL` only, <60s expected. A guard must fail the run if `DATABASE_POOLER_URL` resolves to prod (see WT-4; until then, never point tests at prod).
3. Discord interaction changes: `make run` against a **dev guild** with a dev bot token (see `scripts/smoke.sh` when it lands); guild-scoped registration propagates instantly. Never run the prod bot as a test instance.
4. Web changes: `make run-web` + browser; handlers are httptest-able.

## Task Tracking & Documentation

When completing significant structural/organizational changes (per AGENTS.md history):
- Document folder/file changes (deletions, moves, renames)
- Document command changes (additions, removals, subcommand changes)
- Update this file if it affects agent workflow
