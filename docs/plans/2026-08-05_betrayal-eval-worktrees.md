# Betrayal Bot — Project Evaluation, Worktree Plan & Agent Notes

> **For Hermes/agents:** This is a *planning + evaluation* document. It does three things: (1) reports the jank found in the codebase, (2) defines worktree-sized workstreams (WT-1…WT-8) you can branch off and tackle in parallel, and (3) contains the drafted replacement `AGENTS.md` so agents have correct workflow/context notes before touching the repo.
>
> **Status:** ✅ **Executing** (session 2026-08-05). Progress log at the bottom.

**Goal:** Turn the current "janky but working" state into a maintainable codebase with parallel worktree workflows, a sane local testing story, and a proper admin web panel.

**Context / assumptions:**
- Go 1.23, discordgo + zekroTJA/ken (slash commands), pgx/v5 + sqlc models, Echo + templ + HTMX + Tailwind v4 web panel, zerolog logging w/ DB audit trail, Railway-hosted prod (despite a Fly workflow file).
- `make build`, `go build ./...`, `go vet ./...` all pass today. There are only ~10 test functions, most requiring a live Postgres.
- User pain points (explicit): slow Discord-API testing loop (tear down / spin up the bot), web panel needs upgrades + admin tooling, and worktrees need an env-variable solution (many `.env` values needed per checkout).

---

## Part 1 — Evaluation: What Is Jank (and why)

### A. Repository hygiene & policy violations

| # | Finding | Location | Severity |
|---|---------|----------|----------|
| A1 | **20 MB `betrayal-bot` binary is committed to git** — directly violates the repo's own AGENTS.md rule ("DO NOT commit binary files"). `git ls-files` shows it tracked. | repo root | 🔴 High |
| A2 | Filename typo: `chanenls.go` | `internal/discord/chanenls.go` | 🟡 Low (rename to `channels.go`) |
| A3 | Field typo: `conifg` used as struct field name | `cmd/betrayal-bot/main.go:70,160` | 🟡 Low |
| A4 | Migration file naming is inconsistent: `000016_player_confessional.sql.up.sql` (double `.sql`), `000017_player_immunity.down.sql`, `000018_admin_channels.down.sql` (missing `.sql`) | `internal/db/migration/` | 🟡 Low |
| A5 | Docs sprawl — 5 overlapping docs: `README.md`, `AGENTS.md`, `ADMIN_ANALYSIS.md`, `CHANNEL_QUICK_REFERENCE.md`, `WEB_ADMIN_PLAN.md` repeat the same channel/DB tables with drift (e.g., ADMIN_ANALYSIS says audit retention 365 days, README says 90). | repo root | 🟡 Medium |
| A6 | `.opencode/` tool-specific dir (plans, node_modules, plugin config) is committed to the repo. | `.opencode/` | 🟡 Low — decide keep or ignore |
| A7 | `go.mod` carries unused-ish deps: `lib/pq` (only for migrate CLI, not the app), `go-redis` (only via dgrs indirect). Run `go mod tidy` to verify. | `go.mod` | 🟢 Info |

### B. Actual bugs & broken code (FIXME map)

| # | Finding | Location |
|---|---------|----------|
| B1 | **Likely real bug:** `bot.Identify.Intents = discordgo.PermissionAdministrator` assigns a *permission* constant (`1<<3` = 8) to the *intents* field — value 8 is `IntentGuildEmojisAndStickers`. The bot likely only subscribes to emoji events. Slash commands still arrive (interactions don't need intents), but Ken's internal state (guilds/members/channels) is starved. Verify against prod behavior; likely should be `discordgo.IntentsAll` or an explicit set. | `cmd/betrayal-bot/main.go:136` |
| B2 | **Broken feature:** ability roll path is dead — marked "FIXME: This is broken". `GetRandomAnyAbilityByMinimumRarity` may not exist/misbehave. | `internal/commands/roll/roll.go:293-294` |
| B3 | **Hardcoded prod/test channel ID** for command logging (`testLoggerID = "1108318770138714163"`). Should be a DB row / env var / slash command (there's already a TODO). | `cmd/betrayal-bot/main.go:316` |
| B4 | **Hardcoded game constants**: `defaultCoins = 200`, `defaultItemsLimit = 4`, `defaultLuck = 0` ("TODO: Maybe make these configurable?"). | `internal/commands/inv/create.go:19-24` |
| B5 | **"Unholy" switch chains** for add/remove/set logic per category. | `internal/commands/inv/create.go:163+` |
| B6 | **Categories display is half-broken** ("FIXME: Categories need to be queried/overhauled") — `view` queries category names but the design is incomplete. | `internal/commands/view/view.go:205-220` |
| B7 | **Component builder weirdness** ("FIXME: What the actual hell") — duplicated `clearAll` bool flags threaded through button builders. | `internal/commands/help/player.go:49-73` |
| B8 | Service layer has a function literally named **`Jank()`** with a "one off hack" warning. | `internal/services/inventory/inventory.go:62-66` |
| B9 | Startup jank: logger `Init` called twice (temp logger → re-init with DB), and `Unregister()` called twice "to remove any lingering commands". | `cmd/betrayal-bot/main.go:105-167, 219-243` |
| B10 | `context.TODO()` used in service writes instead of request contexts. | `internal/services/inventory/item.go:15`, `status.go:15` |

### C. Testing situation (your #1 pain point)

- `tests/inventory/` requires a **local Postgres socket at `/tmp/.s.PGSQL.5432`** or it silently skips — so on a normal machine the inventory tests just don't run.
- `tests/database/fuzzy_test.go` connects to **`DATABASE_POOLER_URL` — which is the PRODUCTION Railway pooler** (`roundhouse.proxy.rlwy.net`). Running `go test ./...` with the local `.env` present executes fuzzy queries against the live prod database. 🚨
- `tests/.env` exists on disk with **live production credentials** (Discord bot token, client secret, Railway API token, prod DB URL). It is gitignored (the `.env` pattern matches at any depth), but it is one careless `git add -f` or repo copy away from leaking; those tokens should be treated as compromised and rotated.
- Zero unit tests for: command handlers (ken), services, web handlers, util packages.
- Discord-API testing loop: the only way to exercise a slash command today is to boot the real bot against real Discord (connect, unregister/register 15 global commands, wait for propagation), type the command in a guild, then tear down. No harness, no dev-guild registration, no hot reload.

### D. Web admin panel

**Good bones:** Echo + templ + HTMX + Tailwind v4 ("Dusty Western" theme), `DISABLE_DISCORD=true` web-only mode, `/health` + HTMX partials, roles CRUD (recent), votes dashboard, audit log viewer, Railway redeploy button, vendored htmx, generated templates committed so `go build` works offline.

**Gaps vs. its own plan (`WEB_ADMIN_PLAN.md` "Security Considerations"):**
- ❌ No CSRF middleware (Echo has it built in; must be wired for HTMX).
- ❌ No rate limiting on `/login` or `/admin/redeploy`.
- ⚠️ Single shared `ADMIN_PASSWORD`; `SESSION_SECRET` silently falls back to the admin password (`server.go:50-54` — comment says "not ideal but works").
- ❌ No admin tooling for the actual game state: cycle control, channel-config validation, player management (edit coins/items/status), item/ability/status CRUD (only roles so far).
- ❌ Zero handler tests (Echo `httptest` makes these cheap once DB is abstracted).

### E. Deployment confusion

- `.github/workflows/fly-deploy.yml` deploys to **Fly** on push to main; but the app is **hosted on Railway** (Railway API client + env vars + Dockerfile `ENV_FILE` build arg). One of these is a zombie — confirm which and delete/adjust the other.
- Dockerfile is single-stage, installs templ + tailwind at build, copies `.env` via build arg (secret baked into image layers — prefer Railway env vars).
- No CI for tests/build/lint; Fly deploy is the only pipeline and it's for the wrong platform.

---

## Part 2 — Worktree Workstreams

**Recommended worktree creation flow (see Part 3 for the env piece):**

```bash
# From main, after the WT-3 tooling lands:
git worktree add ../betrayal-wt4 -b wt4-test-infra
scripts/dev-env.sh link        # symlink .env from ~/.config/betrayal/env
make db-up && make mock-migrate-up
```

Each workstream below is sized so it can live on its own branch/worktree without colliding. **Do WT-1 and WT-3 first** — they unblock everything else.

### WT-1 — Repo hygiene & credential rotation 🔴 (unblocks all)

- `git rm --cached betrayal-bot`; add `betrayal-bot`/`bin/` to `.gitignore`; add a `make clean` target. (History keeps the old blob — that's fine, or use `git filter-repo` if you want it gone from history.)
- Rename `internal/discord/chanenls.go` → `channels.go` (update package, no API change).
- Rename `conifg` → `config` in `app` struct.
- Standardize migration filenames for future migrations (document convention; renaming applied migrations is risky — leave existing files, adopt `NNNN_name.up.sql` / `NNNN_name.down.sql` going forward).
- Delete `tests/.env`; **rotate** the Discord bot token, Discord client secret, and Railway API token it contained (they're live credentials that sat in a plaintext file).
- Add a `.env`-leak guard: CI job that fails if any tracked file matches `*_templ.go`... no — fails if a tracked file contains known secret prefixes. Simpler: `git secrets` or a tiny CI grep for `RAILWAY_API_TOKEN=` in tracked files.
- **Verify:** `git status` clean of binaries; `go build ./...`; `git grep -c "testLoggerID"` still 1 (expected).

### WT-2 — Local dev environment (docker-compose Postgres + make targets)

- Add `docker-compose.yml` (or `compose.yaml`) with a single `postgres:16` service, port 5432, volume for persistence, healthcheck. No Redis needed (state is internal).
- Add `.env.example` entries: `DATABASE_URL` (local test DB, distinct from `DATABASE_POOLER_URL` which stays prod) and `MOCK_DATABASE` (used by `make mock-migrate-*`).
- Makefile: replace `grep`-based env extraction with a helper (`ENV_FILE ?= .env`, `env-value = $(shell grep -E '^$1=' $(ENV_FILE) | cut -d= -f2- | tr -d '"')`) or switch migrations to a small `scripts/migrate.sh`.
- Targets: `make db-up`, `make db-down`, `make db-logs`, `make test-up` (db-up + mock-migrate-up).
- **Verify:** `make db-up && make mock-migrate-up && go test ./tests/database/...` runs against *local* Postgres only (confirm by checking the pool URL host).

### WT-3 — Worktree + env tooling (your explicit ask) 🎯

**Problem:** `.env` is gitignored, so every `git worktree add` lands with no config; copying a 20-line secret file per checkout is error-prone and drifts.

**Solution — single source of truth + symlink:**

1. **Canonical env lives outside the repo:** `~/.config/betrayal/env` (gitignored by nature). Contains the full env: prod `DATABASE_POOLER_URL`, `DISCORD_BOT_TOKEN`, Railway vars, `ADMIN_PASSWORD`, `SESSION_SECRET`, plus local `DATABASE_URL` and `MOCK_DATABASE`.
2. **Each worktree's `.env` is a symlink** to that file. godotenv autoload reads `./.env`, so everything (bot, web, tests, Makefile) keeps working untouched.
3. **`scripts/dev-env.sh`** (committed):
   - `link` — create `.env` symlink (or copy `~/.config/betrayal/env` if missing: `cp .env.example ...`).
   - `new-worktree <name> [-b branch]` — `git worktree add`, then `link`, then `make db-up && make mock-migrate-up`.
   - `doctor` — verify symlink target exists, all required keys present (compare against `.env.example`), DB reachable.
4. **Per-worktree overrides:** if one worktree needs a *different* bot token / DB (e.g., testing migrations that could break prod schema), replace the symlink with a real file — documented in AGENTS.md. Never edit the shared file for a one-off.
5. **Guard:** `scripts/dev-env.sh doctor` fails loudly if `DATABASE_POOLER_URL` is unset or if a test run would point at prod (see WT-4 guard).

**Files:** `scripts/dev-env.sh`, `scripts/migrate.sh`, Makefile targets `make env-link`, `make worktree` (wraps `dev-env.sh new-worktree`).
**Verify:** from a fresh `git worktree add`, `make env-link && make run-web` boots web-only mode; `make mock-migrate-up` works.

### WT-4 — Test infrastructure (the "fast local testing" fix) 🔴

Goal: run meaningful tests in seconds against a local DB, and stop any test from ever touching prod.

1. **Test DB isolation:**
   - Tests read `DATABASE_URL` (local) **only**. Add a hard guard in a shared `tests/suite_test.go` helper: if `DATABASE_POOLER_URL` contains the prod host (`roundhouse.proxy.rlwy.net`) → `t.Fatal` ("refusing to run tests against production").
   - Per-suite schema: each suite runs `mock-migrate-up` once (TestMain) + truncates tables between tests (`TRUNCATE ... RESTART IDENTITY CASCADE` on a `tests/truncate.sql` list) instead of relying on a live socket skip.
   - Remove the `/tmp/.s.PGSQL.5432` socket skip in `tests/inventory/suite_test.go`; replace with "fail if `DATABASE_URL` unreachable" (docker compose makes it reachable).
2. **ken command harness (the big win):**
   - Abstract Discord side-effects behind small interfaces in `internal/discord` or a new `internal/testutil`: e.g. `type Responder interface { RespondEmbed(...); RespondMessage(...) }` — command handlers already route through `ctx`; the pragmatic path is to extract the *decision logic* out of handlers into pure functions/services that take `(q *models.Queries, args ...)` and return `(embed, error)`. Then unit-test the pure layer with the local DB and keep ken handlers as thin shells. (Full ken-context mocking is not worth it; ken ships no mock and its Ctx is heavy.)
   - Service layer already exists (`internal/services/inventory`) — extend the pattern to other domains (roll, cycle, vote) and test those.
3. **Web handler tests:** Echo supports `httptest` natively — handlers take `*pgxpool.Pool`, so spin the local DB + `echo.New()` + real routes, hit them with `httptest.NewRequest`. Cover auth (login/logout), `/players`, `/roles` CRUD, `/health`.
4. **CI:** add a `test` job (postgres service container, `make db-up`-equivalent via GitHub Actions `services:` block, `go test ./...`, `go vet`, `go build`).
5. **Discord integration smoke tests stay manual but scripted:** `scripts/smoke.sh` documents the 5-minute manual pass (dev guild, `/healthcheck`, one `/inv create`). Optionally add a `-register-guild <id>` flag to main so commands register **guild-scoped** (instant propagation vs. global) — this alone fixes most of the "slow spin-up" pain for testing new commands.

**Verify:** `go test ./...` green in <60s on a laptop with `docker compose up -d` running; `grep -r "DATABASE_POOLER_URL" tests/` returns nothing.

### WT-5 — Command-layer refactor + bug fixes (B-list)

Do after WT-4 so fixes are testable:
1. Fix B1 (intents) — highest priority; add a test asserting intents bits.
2. Fix B2 (roll ability path) — restore `GetRandomAnyAbilityByMinimumRarity` or route through the same query the item path uses.
3. B3: move log-channel ID to a DB table (migration `000028`) or env var; wire `/channel log` admin subcommand (replaces the TODO at `main.go:308`).
4. B4: game constants → `game_config` table or env vars with the current defaults as fallback.
5. B5/B7: replace `inv create` switch chains with a per-category operation map; clean the help-button component builder.
6. B6: finish or remove the categories field in `/view` ability embeds.
7. B8: rename `Jank()` → `NewManualInventoryHandler()` (or similar) with a real doc comment.
8. B9: single-pass logger init (create pool first, then `logger.Init` once); drop the double `Unregister()`.
9. B10: thread real contexts (`ctx.GetEvent().Context`? no — use `context.WithTimeout(context.Background(), ...)`) through service calls.

**Verify:** `go vet ./...` clean; new unit tests for each fixed path; manual smoke on a dev guild for roll/view/inv.

### WT-6 — Web admin upgrades & admin tooling

1. **Security (from WEB_ADMIN_PLAN.md, never done):**
   - Echo CSRF middleware (token in cookie + HTMX `HX-Request` header handling).
   - Rate limiting on `/login` (e.g. `golang.org/x/time/rate` per-IP or echo middleware) and a confirm-step on `/admin/redeploy`.
   - Fail startup (or refuse to start web) when `SESSION_SECRET` is unset instead of falling back to the password.
2. **Game-state admin pages (the real "admin tooling" ask):**
   - `/cycle` — view/advance/set phase from the web (mirror `/cycle` commands).
   - `/channels` — config validation page: list vote/action/admin/lifeboard channels, mark missing/orphaned (implements the long-standing "Missing Features" list in AGENTS.md: `/admin health` equivalent as a page).
   - `/players/:id/edit` — edit coins, luck, items, statuses, perks from the web (reuses `internal/services/inventory` — avoids duplicating rules).
   - `/items`, `/abilities`, `/statuses` CRUD — generalize the roles CRUD pattern (WT-5's per-category maps make this tractable).
3. **UX polish:** keep the Dusty Western theme (matches your taste profile: warm, non-corporate); mobile-first pass on tables (players/audit) — they're the pages you'd check from a phone mid-game.
4. Handler tests (from WT-4.3) for every new route.

**Verify:** manual pass on `make run-web` (no Discord needed): login → edit a role → change cycle → confirm audit page reflects it; `go test ./tests/web/...` (new) green.

### WT-7 — Docs consolidation + AGENTS.md rewrite

- Delete `ADMIN_ANALYSIS.md` + `CHANNEL_QUICK_REFERENCE.md` + `WEB_ADMIN_PLAN.md` once their unique content is absorbed (channel quick-ref → AGENTS.md; web plan → `internal/web/README.md` + roadmap notes; analysis → AGENTS.md "Missing Features").
- Replace `AGENTS.md` with the draft in Part 4 below.
- Update `README.md` to point at AGENTS.md for workflow and drop its drift (e.g., audit retention 90 vs 365).
- **Verify:** grep the deleted filenames across the repo — no dangling references.

### WT-8 — CI & deployment clarity

- ~~Decide Railway vs Fly~~ **DONE: Fly removed** — `fly-deploy.yml` deleted 2026-08-05 (user confirmed "old and no longer in use"). Railway is canonical; the in-app redeploy button already exists.
- Add CI: `lint` (gofmt check, `go vet`), `test` (from WT-4), `build` (`make build`), `secret-leak` guard (from WT-1).
- Dockerfile: multi-stage (builder + runtime), drop the `ENV_FILE` build arg (use runtime env vars), keep templ+tailwind generation in the builder stage.
- **Verify:** CI green on a PR; `docker build` succeeds without `.env`.

---

## Part 3 — The Worktree + Env Solution (detail)

**The core insight:** gitignored `.env` + `git worktree` = every checkout is born unconfigured. Symlinking to one canonical env file (outside the repo) makes all worktrees instantly runnable, and a real-file override handles the rare "this worktree needs different creds" case.

```
~/.config/betrayal/env            ← canonical (prod + local keys), chmod 600
../betrayal-wt4/.env → ~/.config/betrayal/env   (symlink created by dev-env.sh)
../betrayal-wt6/.env → ~/.config/betrayal/env
```

**Env key inventory** (from `.env.example` + `main.go`):
| Key | Used by | Notes |
|-----|---------|-------|
| `DISCORD_BOT_TOKEN` / `CLIENT_ID` / `CLIENT_SECRET` | bot auth, OAuth | rotate after WT-1 |
| `DATABASE_POOLER_URL` | prod DB pool (app) | tests must NEVER use this |
| `DATABASE_URL` | **local** test DB (new) | WT-2 adds |
| `MOCK_DATABASE` | `make mock-migrate-*` | local DB for migrations |
| `WEB_PORT`, `ADMIN_PASSWORD`, `SESSION_SECRET` | web panel | require SESSION_SECRET after WT-6 |
| `RAILWAY_API_TOKEN`, `RAILWAY_*_ID` | redeploy button | rotate after WT-1 |
| `ENVIRONMENT` | logger mode (`local` vs `production`) | tests force `local` |
| `EVIL_ROLES_CSV`, `GOOD_ROLES_CSV`, `NEUTRAL_ROLES_CSV`, `ITEM_CSV` | `/setup` data entry | Google Sheets URLs — lives in `setup.go:251` "TODO: find me a better home" |

---

## Part 4 — Drafted `AGENTS.md` (the "updated notes for agents")

> Apply as part of WT-7, or tell the agent to apply now. This replaces the current AGENTS.md entirely (its channel quick-reference is preserved below).

```markdown
# Agent Guidelines for Betrayal Bot

## What this bot is
Discord game-management bot for "Betrayal" (battle-royale game). Go 1.23,
discordgo + zekroTJA/ken (slash commands), pgx/v5 + sqlc (internal/models),
Echo + templ + HTMX + Tailwind v4 web admin panel, zerolog logging + DB audit
trail. Hosted on Railway (prod). Do NOT trust README/old docs over this file.

## Build & run
- `make run` — full bot (Discord + web). Requires `.env` (see Worktrees & Env).
- `make run-web` — web panel only (`DISABLE_DISCORD=true`). Fastest way to
  iterate on the admin UI; no Discord needed.
- `make build` — templ generate + tailwind build + go build to ./bin/.
- `make generate` — templ + tailwind only (needed after editing *.templ or input.css).
- Hot reload: `air` is supported (`.air.toml` is gitignored); `templ generate --watch`
  + `tailwindcss --watch` for template/CSS iteration.
- Tests: `go test ./...` — REQUIRES local Postgres (`make db-up` first). Tests must
  never touch the production DB; a guard fails the run if `DATABASE_POOLER_URL`
  resolves to prod.
- Migrations: `make migrate-up/down` (prod via DATABASE_POOLER_URL),
  `make mock-migrate-up/down` (local via MOCK_DATABASE). Never run migrate-up
  against prod unless you mean it.

## Worktrees & env (READ FIRST)
- `.env` is gitignored; each worktree gets a SYMLINK to the canonical env file
  at `~/.config/betrayal/env`. Run `scripts/dev-env.sh link` after creating a
  worktree (or `make env-link`).
- New worktree: `scripts/dev-env.sh new-worktree <name>` → creates worktree,
  links env, brings up local DB + migrations.
- Overrides: replace the symlink with a real `.env` file for one-off creds;
  NEVER edit the shared file for a one-off, and never `git add -f` any `.env`.
- Required keys: see `.env.example`. `ENVIRONMENT=local` in dev (console logs);
  `production` writes logs to the DB.
- If `dev-env.sh doctor` fails, fix the reported key before running anything.

## Repo layout
- `cmd/betrayal-bot/main.go` — wiring: config from env, DB pool, Ken init,
  command registration list, web server. Add new commands HERE.
- `internal/commands/{name}/` — ken command packages (struct implements
  `ken.Command` + `Initialize(*pgxpool.Pool)`; `var _ ken.SlashCommand = (*X)(nil)`).
- `internal/services/` — reusable game logic (inventory exists; keep new rules
  here so they're unit-testable WITHOUT Discord).
- `internal/models/` — sqlc-generated query code; edit `internal/db/query/*.sql`
  and regenerate (do not hand-edit *.sql.go).
- `internal/db/migration/` — golang-migrate files; name as `NNNN_name.up.sql` /
  `NNNN_name.down.sql` (NOTE: old files 000016-000018 have inconsistent names;
  do not rename applied migrations).
- `internal/discord/` — embed/error/component helpers (chanenls.go is a legacy
  typo filename; new files use correct spelling).
- `internal/web/` — Echo server, handlers, middleware, railway client,
  templates (templ). Static assets in `web/static/` (tailwind input/output.css,
  vendored htmx.min.js).
- `tests/` — testify suites. DB suites need local Postgres; logger suites are
  unit tests. Web handler tests live here too (httptest).

## Command inventory (registered in main.go)
| Command | Purpose | Admin? |
|---------|---------|--------|
| /inv | inventory mgmt: ability/item/coin/status/perk/alignment/role/immunity/luck/death/notes + create | most subcommands admin |
| /roll | rolls: item/ability rarity rolls, player choice, luck, event rolls | hybrid |
| /action | submit game action (confessional) | player |
| /view | view role/ability/item/status details with buttons | player |
| /buy | purchase item for player | player |
| /channel | channel config: admin/vote/action/lifeboard/confessionals | admin |
| /help | help embeds (player + admin) | both |
| /vote | cast votes (funnel channel) | player |
| /setup | generate role list from CSV data entry | admin |
| /echo | ping/debug | admin |
| /list | list roles/items/statuses/etc | player |
| /search | fuzzy search abilities/items/statuses | player |
| /healthcheck | bot health | admin |
| /cycle | current/next/set phase + broadcast to confessionals/funnels/alliances | admin |
| /tarot | tarot draws (deterministic/per-user/guild-deck/random) | both |

Admin roles (role.go): Host, Co-Host, Bot Developer — check with
`discord.IsAdminRole(ctx, discord.AdminRoles...)`, respond with
`discord.NotAdminError(ctx)`.

## Web admin panel
- Routes (internal/web/server.go): `/login`, `/` dashboard, `/health`,
  `/players` + `/players/:id`, `/votes`, `/roles` CRUD, `/admin/audit`,
  `/admin/redeploy` (Railway). All protected by session auth except /login,/health.
- Editing templates requires: `templ generate` (commits _templ.go) and
  `tailwindcss -i web/static/css/input.css -o web/static/css/output.css --minify`.
  Both run via `make generate`. CSS source is Tailwind v4 (`@import "tailwindcss"`
  in input.css; theme tokens in `@theme` — "Dusty Western" palette).
- Theme: warm, non-corporate, mobile-first. Keep it that way.
- Security TODOs (do before adding public-facing routes): CSRF middleware,
  login rate limiting, require SESSION_SECRET (no password fallback).

## Known jank register (fix under WT-5, don't perpetuate)
- main.go:136 sets Intents from a permission constant — verify/fix intents.
- roll.go:293 ability roll path broken (FIXME).
- view.go:213 categories FIXME; help/player.go:49 button-builder FIXME.
- inv/create.go: hardcoded game constants + "unholy" switch chains.
- main.go:316 hardcoded command-log channel ID (migration 000028 planned).
- inventory service `Jank()` is a documented hack — prefer NewInventoryHandler.
- Logger init + Ken Unregister happen twice at startup (cleanup planned).

## Deployment
- Prod = Railway (env-driven). `.github/workflows/fly-deploy.yml` is legacy —
  do not extend it; deployment migration is WT-8.
- Dockerfile: single-stage today; build needs no .env (runtime env only).
- Never commit binaries (`betrayal-bot`, `bin/`) or `.env` files.

## Testing workflow (make it fast)
1. `make db-up` (docker compose postgres) once per machine/worktree.
2. `go test ./...` — unit + DB suites, <60s, local-only.
3. For Discord interaction changes: `make run` against the DEV GUILD using a
   dev bot token (see `scripts/smoke.sh`); guild-scoped registration lands
   instantly. Never run the prod bot as a test instance.
4. Web changes: `make run-web` + browser; handlers are httptest-able.
```

---

## Part 5 — Suggested Execution Order & Handoff

1. **Now (this session, if you approve):** apply WT-1 hygiene fixes + write `scripts/dev-env.sh` + draft AGENTS.md replacement (WT-7 partial) — small, unblocks everything.
2. **WT-3 → WT-2 → WT-4** — the local-dev foundation; worktrees become cheap and tests stop touching prod.
3. **WT-5** (bug fixes) and **WT-6** (web/admin) in parallel worktrees once tests exist to catch regressions.
4. **WT-8** when deployment is confirmed Railway-only.

**Worktree collision map:** WT-1 touches root/main.go lightly; WT-2/3 touch Makefile + scripts (merge WT-2 and WT-3 into ONE worktree to avoid Makefile conflicts); WT-4 tests/ + CI; WT-5 internal/commands; WT-6 internal/web; WT-7 docs; WT-8 CI/Docker. Only real conflict: WT-4 and WT-8 both edit CI config — sequence them.

---

## Risks, Tradeoffs & Open Questions

1. **Credential rotation** (WT-1): the tokens in `tests/.env` are live. Rotating the Discord bot token requires updating the Railway env + any local env files. Do it once, early.
2. **ken testing depth:** full ken-context mocking is rejected (no official mock; heavy Ctx). The bet is "extract logic to services, test services." If you'd rather have integration-style tests that boot ken with a fake session, that's a bigger WT-4 variant — speak up.
3. **Fly vs Railway:** need a decision. Assumption is Railway (all live config says so); Fly workflow file is likely dead.
4. **Prod schema vs test schema drift:** tests use the real migrations against local DB — good. But WT-5's new migrations (log channel, game config) must be written carefully since prod applies them via `make migrate-up`.
5. **`.opencode/` in git:** keep (plans are useful history) or ignore? Minor.
6. **`go mod tidy`:** safe to run? It may drop lib/pq if truly unused — verify migrate CLI still works from the Makefile path.
7. **Test DB in CI:** GitHub Actions `services: postgres` is the cheap path; testcontainers is the heavier alternative. Recommend services.

---

## Progress (2026-08-05)

All five workstreams merged into `main`, verified, branches removed:
WT-4 (test infra), WT-5 (command fixes), WT-6 (web admin), WT-7 (docs),
WT-9 (Dark Obsidian Glass UI theme). Post-merge cleanup: Makefile `env-value`
helper, web suite joined `testutil.Bootstrap` (advisory lock + prod guard),
CSRF-aware web tests, roll pin test updated post-fix, `web.New` 2-value
signature, `Jank()` → `NewManualInventoryHandler` fallout. `go vet` +
`go test ./...` green on merged main.

**Before deploying merged main to Railway:** run `make migrate-up` on prod
(migrations 000028 `command_log_channel`, 000029 `game_config`) and set
`SESSION_SECRET` (≥32 bytes) in the Railway env — `web.New` refuses to start
without it. `make run-web` connects to `DATABASE_POOLER_URL` (prod) unless
overridden with the local `DATABASE_URL`; the WT-6 pages added after WT-9's
theme pass (items/statuses/abilities/cycle/channels) need a glass-theme polish.
