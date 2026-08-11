# Agent Guidelines for Betrayal Bot

## What this bot is

Discord game-management bot for "Betrayal" (battle-royale game). Go 1.23, discordgo + zekroTJA/ken (slash commands), pgx/v5 + sqlc (`internal/models/`), Echo JSON APIs + SvelteKit static web admin panel, zerolog logging with DB audit trail. Hosted on **Railway** (prod). The legacy Fly workflow was removed (2026-08) — do not reintroduce it.

## Build & Run

**Prereqs (one-time per machine):** Node.js/npm for the SvelteKit frontend, golang-migrate CLI with the postgres driver (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1`), and `$HOME/go/bin` on PATH.

- **Full bot** (Discord + web): `make run` — requires `.env` (see Worktrees & Env).
- **Web panel only**: `make run-web` (sets `DISABLE_DISCORD=true`) — fastest way to iterate on the admin UI, no Discord needed.
- **Build**: `make build` → validates migration pairs, builds the SvelteKit static output, then `go build` to `./bin/`.
- **Release preflight**: run `make install-hooks` once per checkout to enable the tracked pre-push guard; it runs `make test-release` (migration-pair validation plus the migration recovery tests).
- **Generate assets**: `make generate` — required after editing frontend SvelteKit files.
- **Hot reload**: use the SvelteKit Vite dev server for frontend iteration; production remains static output served by Go.
- **Tests**: `go test ./...` — REQUIRES local Postgres (`make db-up` first; migrations are applied by the test bootstrap itself). Tests must never touch the production DB — a hard guard enforces it (see Testing Workflow).
- **CI**: `.github/workflows/test.yml` — postgres service container + `go vet ./...` + `go test ./...` + `make build` on push/PR.
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
- `internal/services/` — reusable game logic (inventory, roll, cycle, vote, **roledraft**). Keep new rules here so they're unit-testable WITHOUT Discord.
- `internal/models/` — sqlc-generated query code. Edit `internal/db/query/*.sql` and regenerate; do NOT hand-edit `*.sql.go`.
- `internal/db/migrate/migrations/` — golang-migrate files, EMBEDDED via `//go:embed` (runner in `internal/db/migrate/`; the web panel's `/admin/migrations` page and the Makefile both operate on these). Name new ones `NNNN_name.up.sql` / `NNNN_name.down.sql` (legacy files 000016–000018 have inconsistent names; do not rename applied migrations).
- `internal/discord/` — embed/error/component/channel helpers. (`channels.go` was renamed from the legacy `chanenls.go` typo.)
- `internal/web/` — Echo server, JSON API handlers, session middleware, Railway client, and embedded SvelteKit output (see `internal/web/README.md`).
- `tests/` — testify suites. DB suites use the shared bootstrap in `tests/testutil` (production guard + migrations + truncation); web handler tests drive the real Echo routes with httptest.
- `scripts/dev-env.sh` — worktree/env tooling (see above).

## Command Inventory (registered in `main.go`)

| Command | Purpose | Admin? |
|---------|---------|--------|
| `/inv` | inventory mgmt: ability/item/coin/status/perk/alignment/role/immunity/luck/death/notes + create | most subcommands |
| `/roll` | rolls: item/ability rarity, player choice, luck, event rolls | hybrid |
| `/action` | submit game action (confessional) | player |
| `/view` | view role/ability/item/status details with buttons | player |
| `/buy` | purchase item for player | player |
| `/channel` | channel config: admin/vote/action/lifeboard/confessionals/log | admin |
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

## Channel Configuration (quick reference)

Five channel types drive the game; all are configured via `/channel` (admin-only). Code lives in `internal/commands/channels/{channels,admin,vote,action,lifeboard}.go`; generated queries in `internal/models/*.sql.go`.

| Type | Cardinality | Purpose | Command | DB table (migration) |
|------|-------------|---------|---------|----------------------|
| Admin | multiple | Whitelist where `/inv` works outside confessionals (`inv/get.go:41`) | `/channel admin [add/list/delete] [channel]` | `admin_channel` (000018) |
| Vote | single | Funnel where players submit votes; target of `/cycle` broadcasts | `/channel vote [update/view] [channel]` | `vote_channel` (000019) |
| Action | single | Funnel where players submit actions; target of `/cycle` broadcasts | `/channel action [update/view] [channel]` | `action_channel` (000020) |
| Lifeboard | single | Pinned player status board — alive players A–Z, then dead A–Z, EST footer | `/channel lifeboard set [channel]` | `player_lifeboard` (000021) |
| Confessional | one per player | Private player↔admin channel; created during game setup, NOT via `/channel` | `/channel confessionals` (view only) | `player_confessional` (000016) |

**Setup order (new game):**
1. Create a confessional per player (outside `/channel`).
2. `/channel vote update #funnel`, then `/channel action update #funnel`.
3. `/channel admin add #ops` (repeatable for multiple).
4. `/channel lifeboard set #status-board`.
5. Verify: `/channel confessionals`, `vote view`, `action view`, `admin list`.

**Cycle broadcast flow** (`cycle.go`): `/cycle next|set` messages go to all confessionals + vote channel + action channel + every channel in the `alliances` category (via `discord.GetChannelsWithinCategory`).

**Error recovery:**
- Vote/action "not found" → `/channel vote view` / `action view` to confirm they're set.
- Confessional missing cycle messages → `/channel confessionals` and re-create any missing ones.
- Lifeboard stale → re-run `/channel lifeboard set #channel` (deletes + rebuilds + re-pins).
- Admin commands failing → confirm the user holds Host, Co-Host, or Bot Developer.

## Web Admin Panel

- Routes (`internal/web/server.go`): `/login`, `/` dashboard, `/health`, `/players` + `/players/:id` + `/players/:id/edit`, `/cycle` (+ `/cycle/advance`, `/cycle/set`), `/channels` (validation + mutations), `/setup` (role-pool generator), `/votes`, `/roles` CRUD, `/items` `/abilities` `/statuses` CRUD (search/create/detail/update/delete), `/sync` (spreadsheet sync: preview/apply/sources), `/admin/audit`, `/admin/migrations` (embedded runner: up/rollback with confirmation), `/admin/redeploy` (Railway). Session-auth protected except `/login` + `/health`. Full route table: `internal/web/README.md`.
- Security: Echo CSRF (double-submit cookie for JSON requests), login + redeploy rate limiting, and password-derived signed sessions.
- **The app defaults to `DATABASE_POOLER_URL` (production)** — `ENVIRONMENT=local` is an explicit local-development opt-in. The panel's write routes (/cycle, player edit, catalog CRUD) therefore hit the LIVE game unless local mode is explicitly selected.
- Frontend changes are made under `frontend/`; run `make generate` and commit the generated `internal/web/ui/dist` output.
- Theme: dark, atmospheric obsidian-glass (game theme is "mirrors"), **mobile-first** — preserve this.
- Handler tests live in `tests/web/` (httptest + local Postgres via `DATABASE_URL`; they seed + clean up their own data).

## Known Jank Register

**WT-5 landed 2026-08-05** (`wt5-command-fixes`): the B-list is fixed —

- Gateway intents: `gatewayIntents()` returns `discordgo.IntentsAllWithoutPrivileged`
  (was the `PermissionAdministrator` permission constant, value 8 = emoji intent);
  asserted by `cmd/betrayal-bot/main_test.go`.
- `roll` ability path: `GetRandomAnyAbilityByRarity` had invalid SQL (`==`),
  and both any-ability roll queries lacked `order by random() limit 1`;
  `GetRandomAnyAbilityByMinimumRarity` now also excludes `UNIQUE` (item parity).
  Covered by `tests/database/roll_test.go`.
- Command-log channel: hardcoded ID replaced by `command_log_channel` table
  (migration 000028) + `/channel log update|view|remove`; `logHandler` is now an
  `*app` method that reads the configured channel and skips when unset.
- `/inv create` defaults (coins 200 / item limit 4 / luck 0) live in the
  `game_config` table (migration 000029, seeded; constants are the fallback).
  `CreatePlayer` now sets `item_limit` from config. `roleOpsByRole` map replaced
  the switch chains (and fixed magician's mislabeled Lucky status + the
  succubus/cultist non-existent status names). Covered by
  `internal/commands/inv/create_test.go`.
- `help/player.go` button-builder FIXME gone (no more dead `clearAll` flags);
  `/view ability` categories only render when present.
- `inventory.Jank()` → `inventory.NewManualInventoryHandler()`.
- Logger initialized once (pool first); double Ken `Unregister()` removed;
  web-only shutdown no longer nil-panics on `betrayalManager`.
- Service writes (`item.go`, `status.go`) use a 10s bounded context via `dbCtx()`.

Still open (don't perpetuate): `inv/create.go` placeholder embed copy
("Idk finished inventory lol"), `roll` `luckTable` is dead (unregistered).

## Missing Features (roadmap)

Documented gaps from the 2026-08 admin analysis (tracked under WT-5/WT-6 — don't build ad-hoc versions):
- No `/admin health` or `/admin status` command: can't verify configured channels still exist in Discord, detect orphaned confessionals, or check configuration completeness before a game starts.
- No channel validation / recovery tooling for channels deleted mid-game (error paths above are manual).
- Web panel lacks game-state admin pages (cycle control, channel-config validation, player edit, item/ability/status CRUD) — see `internal/web/README.md` and the WT-6 workstream in `docs/plans/`.

## Deployment

- Prod = **Railway** (env-driven; in-app redeploy button via `internal/web/railway`). `.github/workflows/fly-deploy.yml` was deleted 2026-08.
- Dockerfile: builds the SvelteKit static output in a Node stage, then compiles the Go server; needs no `.env` at build (runtime env only).
- Never commit binaries (`betrayal-bot`, `bin/`) or `.env` files — `make clean` removes build artifacts.

## Testing Workflow (keep it fast)

1. `make db-up` (docker compose postgres) once per machine/worktree. `go test` applies migrations itself — no `make mock-migrate-up` needed first.
2. `go test ./...` — unit + DB suites against LOCAL `DATABASE_URL` only, <60s expected (measured ~5s). The shared bootstrap (`tests/testutil`, every DB suite's `TestMain`) enforces:
   - **Production guard**: fails the run if `DATABASE_URL` is unset or doesn't resolve to localhost; fails if `DATABASE_POOLER_URL` equals `DATABASE_URL`; then **strips `DATABASE_POOLER_URL` from the test process** so no test can ever route a connection through the prod Railway pooler.
   - **Isolation**: an advisory lock serializes DB test packages sharing the local Postgres, and all tables are truncated between tests (`TRUNCATE ... RESTART IDENTITY CASCADE`, `game_cycle` re-seeded to Day 0).
   - **No silent skips**: unreachable `DATABASE_URL` is a hard failure, not a skip (the old `/tmp/.s.PGSQL.5432` socket skip is gone).
3. Command logic lives in `internal/services/{roll,cycle,vote}` (ken handlers are thin shells) — `tests/{roll,cycle,vote}` unit-test it against the local DB; `tests/web` drives real Echo routes with httptest (login, health, players, roles CRUD).
4. Discord interaction changes: `make run` against a **dev guild** with a dev bot token (see `scripts/smoke.sh` when it lands); guild-scoped registration propagates instantly. Never run the prod bot as a test instance.
5. Web changes: `make run-web` + browser; handlers are httptest-able.

## Non-negotiable engineering gate

- Use strict test-driven development for bug fixes and features: write the
  regression test first, run it and confirm it fails for the intended reason,
  implement the smallest root-cause fix, then run the focused test and full
  verification suite.
- Never push a production change while any test, vet check, build, migration
  check, or required verification is failing. A deploy must fail before serving
  traffic rather than start successfully and return unexplained 500s.
- For web/database changes, test the real request or database seam whenever
  practical; a generic error-page assertion is not sufficient.

## Task Tracking & Documentation

When completing significant structural/organizational changes (per AGENTS.md history):
- Document folder/file changes (deletions, moves, renames)
- Document command changes (additions, removals, subcommand changes)
- Update this file if it affects agent workflow
