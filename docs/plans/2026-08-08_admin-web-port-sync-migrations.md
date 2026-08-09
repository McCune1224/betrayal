# Betrayal Admin → Web Port, Migrations UI & Spreadsheet Sync UI — Implementation Plan

> **For Hermes:** Use subagent-driven-development to implement this plan task-by-task, in worktrees off `main`, with spec-compliance + code-quality review per task and headless verification (`go test ./...`, `make build`) — no game-window or Discord run-around.

**Goal:** Move the admin surface of Betrayal onto the web panel: finish porting every admin slash command (player management, channel config, game setup), add a **database migrations UI** (apply/rollback/status from the browser), and give the **spreadsheet sync** (currently `cmd/data-entry`) a full web UI with fetch → **diff preview → validate → apply** flow. Discord slash commands become an archived layer for end users; admin commands stay registered but are no longer maintained.

**Architecture:** Three shared service layers extracted so both the bot and the web panel consume the same logic: `internal/services/datasync` (CSV parse + diff + apply, replacing the `cmd/data-entry` guts), `internal/db/migrate` (embedded golang-migrate runner), and `internal/services/roledraft` (role-pool generation, extracted from `/setup`). The web panel gets four new/expanded page areas — Sync, Migrations, Setup, and extended Player/Channel pages — all following the existing Echo + templ + HTMX + Tailwind v4 mobile-first "obsidian glass" patterns.

**Tech Stack:** Go 1.23, pgx/v5 + sqlc (`internal/models`), golang-migrate v4.17.1 (already in go.mod), Echo v4 + templ (pin v0.3.960) + HTMX 2 + Tailwind v4.1.2, zerolog. Tests: testify suites in `tests/` (local Postgres only, `tests/testutil` guard).

---

## Part 1 — Current State & Gap Analysis

### 1a. Admin command inventory → web port status

| Command | Subsurface | Web status | Gap to close |
|---|---|---|---|
| `/inv` | create player, coins, luck, immunity, death, notes, alignment, role assign, item/ability/status/perk add-remove, item limit | player edit page has items/abilities/statuses/perks + stats; catalog CRUD exists for items/abilities/statuses; roles CRUD exists | coins, luck, immunity, alive/dead, notes, alignment, role assign, item limit, **create player** |
| `/cycle` | advance/set phase | ✅ `/cycle` advance + set | none (no Discord broadcast by design) |
| `/channel` | admin add/list/delete, vote update/view, action update/view, lifeboard set, log update/view/remove, confessionals view | `/channels` **validation-only** page | **mutations**: set vote/action/lifeboard channel, admin channel add/remove, log channel set/remove |
| `/setup` | role-pool generation (player count + deceptionist count) | ❌ none | new `/setup` page + `roledraft` service |
| `/roll` (admin half) | admin-forced rolls for players | ❌ none | optional (see Open Questions) |
| `/echo`, `/healthcheck`, `/help admin` | debug/info embeds | `/health` exists | skip — informational; stays archived in Discord |
| `/buy`, `/vote`, `/action`, `/view`, `/list`, `/search`, `/tarot` | **player** commands | votes read-only page | stay in Discord (end-user surface) |

**Bot policy change:** admin commands stay registered as an **archive**. No parity work on them going forward. Update `AGENTS.md` + `internal/web/README.md` to say so.

### 1b. Spreadsheet sync (`cmd/data-entry/main.go`, 547 lines)

- Four CSV sources from env: `GOOD_ROLES_CSV`, `EVIL_ROLES_CSV`, `NEUTRAL_ROLES_CSV`, `ITEM_CSV` (Google Sheets CSV exports). `.env.example` doesn't even document them.
- `SyncRolesCsv` parses variable-length "chunks" (role header row → ability rows → `Passives:` → perk rows) and **creates** roles/abilities/perks + category & role links. Unique violations are handled by fuzzy-lookup + link-existing (so re-runs don't hard-fail).
- `SyncItemsCsv` **creates** items + category links.
- Problems: **insert-only** (sheet edits never propagate to existing rows), **no preview/diff** (fires straight into prod DB), console-only output, `cmd/` binary is not embeddable in the web process.
- **Env note:** the CLI reads `DATABASE_POOLER_URL` (prod). Any UI reuse must respect the existing `make run-web` prod-pooler hazard.

### 1c. Migrations

- golang-migrate v4.17.1; 29 migrations in `internal/db/migration/`; applied via Makefile/CLI only. The main binary's `migrate/v4/source/file` import is a no-op (used by tests via `tests/testutil` `migrateUp`, which reads from disk).
- No web visibility, no in-process runner, no rollback from the panel. Migrations are the one operational task with zero UI today.

---

## Part 2 — Proposed Approach (Phases)

- **Phase 0 — Foundations:** extract `internal/services/datasync` (parse → plan → apply) + `internal/db/migrate` (embedded runner); migration `000030` adds `sync_source` + `sync_run` tables; sqlc additions for exact-name lookups.
- **Phase 1 — Sync UI (`/sync`):** sources list (editable URLs), Fetch & Preview → per-source diff (new/update/skip with old→new values) → Apply (transactional + audit row). `cmd/data-entry` becomes a thin CLI wrapper over the same service (archive mode).
- **Phase 2 — Migrations UI (`/admin/migrations`):** status table (version, name, applied/dirty), Apply Pending / Rollback N with confirmation; prod-pooler guard (typed confirmation for destructive ops).
- **Phase 3 — Finish admin port:** player-edit gaps, channel-config mutations, `/setup` role-pool page.
- **Phase 4 — Navigation, docs, cleanup.**

Each phase lands as its own PR/worktree; every task ends with `go test ./...` + `make build` green and a commit.

---

## Phase 0 — Foundations

### Task 0.1: Migration `000030` — sync tracking tables

**Files:**
- Create: `internal/db/migration/000030_sync_source.up.sql` / `.down.sql`
- Create: `internal/db/migration/000031_sync_run.up.sql` / `.down.sql`

**000030_sync_source.up.sql**
```sql
CREATE TABLE IF NOT EXISTS sync_source (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,          -- 'good_roles' | 'evil_roles' | 'neutral_roles' | 'items'
    kind       TEXT NOT NULL,                 -- 'roles' | 'items'
    alignment  TEXT NOT NULL DEFAULT '',      -- GOOD/EVIL/NEUTRAL or '' for items
    url        TEXT NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
**000031_sync_run.up.sql** — audit trail for sync executions:
```sql
CREATE TABLE IF NOT EXISTS sync_run (
    id            BIGSERIAL PRIMARY KEY,
    source_id     INTEGER REFERENCES sync_source(id) ON DELETE CASCADE,
    source_name   TEXT NOT NULL,
    status        TEXT NOT NULL,              -- 'preview' | 'applied' | 'failed'
    action_counts JSONB NOT NULL DEFAULT '{}',-- {"created":3,"updated":5,"skipped":40}
    run_by        TEXT NOT NULL DEFAULT '',   -- web session user or 'cli'
    error_message TEXT NOT NULL DEFAULT '',
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at   TIMESTAMPTZ
);
```

**Step 2 — seed + verify**
Run `make mock-migrate-up` locally; confirm `docker exec betrayal-postgres psql -U postgres -d betrayal -c '\d sync_source'`.

**Step 3 — Commit** `feat(db): add sync_source and sync_run tables`

### Task 0.2: `internal/db/migrate` — embedded migration runner

**Files:**
- Create: `internal/db/migrate/migrate.go`
- Modify: `cmd/betrayal-bot/main.go` (construct + hold the service; pass to web.Config)

```go
// Package dbmigrate runs golang-migrate over the EMBEDDED migration files.
package dbmigrate

import (
    "embed"
    "errors"
    "fmt"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/source/iofs"
    "github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Runner wraps a migrate.Migrate bound to the given pool's DB.
type Runner struct{ m *migrate.Migrate }

func New(pool *pgxpool.Pool) (*Runner, error) {
    src, err := iofs.New(migrationFS, "migrations")
    if err != nil { return nil, err }
    conn, err := pool.Acquire(ctx...) // need a *sql.DB-backed conn for migrate
    ...
}
```

**Design notes / pitfalls:**
- golang-migrate's postgres driver wants `*sql.DB` or a `database/sql`-compatible conn — either use `github.com/jackc/pgx/v5/stdlib` (`stdlib.OpenDBFromPool(pool)`) or open a separate `sql.Open("pgx", dsn)` connection in the runner (preferred: keep it independent of the pool, use `DATABASE_POOLER_URL`/`DATABASE_URL` from env at construction). Simpler and matches Makefile behavior.
- **Move (don't copy) the SQL files**: relocate `internal/db/migration/*.sql` → `internal/db/migrate/migrations/*.sql` so `embed` picks them up, and update `tests/testutil/testutil.go` `migrateUp` to `iofs` over the embed (or keep file:// path but point at the new dir). One source of truth — **no duplication** (DRY). Update Makefile `migrate-up/down` paths to the new dir.
- API surface: `Status() ([]MigrationStatus, error)` (version, dirty, applied_at), `Up() error`, `DownSteps(n int) error`, `Version() (uint, bool, error)`. Never expose raw `Force` in the UI.

**Step — tests**
Create `internal/db/migrate/migrate_test.go` (uses local `DATABASE_URL`; bootstrap a throwaway schema — e.g. migrate to a temp version, assert `Version()`): run `go test ./internal/db/migrate/ -v`. Expected: PASS.

**Step — commit** `feat(db): embedded golang-migrate runner`

### Task 0.3: `internal/services/datasync` — extract + upgrade the CSV logic

**Files:**
- Create: `internal/services/datasync/parse.go` (from `cmd/data-entry/main.go` — `parseRoleChunk`, `parseAbility`, roles/items CSV reading; keep exact column semantics + `∞` charges + `*`/`^` type markers + rarity switch)
- Create: `internal/services/datasync/plan.go` — diff engine
- Create: `internal/services/datasync/apply.go` — transactional apply
- Create: `internal/services/datasync/sources.go` — `sync_source` CRUD (List/Set/Enable) + startup seeding from env
- Delete: `cmd/data-entry/main.go` guts; rewrite as thin CLI: `cmd/data-entry/main.go` now = flag/env-driven call into `datasync.FetchAllAndApply` (keeps the archive workflow identical: same env vars, same prod-pooler behavior).

**Diff engine shape (plan.go):**
```go
type Action string // "create" | "update" | "skip"
type Entry struct {
    Source    string
    Kind      string // "role" | "item" | "ability" | "perk"
    Name      string
    Action    Action
    Changes   []string // human-readable "description: X → Y"
    OldValue  string
    NewValue  string
}
type Plan struct {
    Source  string
    Entries []Entry
    Counts  map[Action]int
}
```
- Matching by **exact name** (new sqlc queries, Task 0.4): `GetRoleByName`, `GetItemByName`, `GetAbilityByName`, `GetPerkByName`, `GetCategoryByName`. Replace the fuzzy lookups used for dedupe with exact ones (fuzzy stays only for the `view`/`search` UX).
- Roles: name match → compare description/alignment → `update`; else `create` (abilities/perks become entries too, with their own create/update + link adds). Items: name match → compare description/rarity/cost → `update`; else `create`.
- `apply.go` runs in **one transaction per source** (pgx `pool.Begin`), inserts a `sync_run` row on completion; on error rolls back + writes a `failed` run. Link logic (role↔ability, role↔perk, category joins) uses `ON CONFLICT DO NOTHING`-style guards (existing `Create*Join` calls; add unique constraints if missing — verify in Task 0.4).

**Step — tests**
- `internal/services/datasync/parse_test.go` — feed a canned CSV (role chunk with 2 abilities + passives, item rows, `∞`, `*`/`^`, unknown rarity) → assert parsed structs. No DB needed.
- `internal/services/datasync/plan_test.go` — seed 1 existing role/item via `tests/testutil` bootstrap; diff a CSV containing (a) same name/desc → skip, (b) changed desc → update, (c) new name → create. Assert `Plan.Counts`.
- `internal/services/datasync/apply_test.go` — apply a plan to local DB; assert rows + `sync_run` row.
Run: `go test ./internal/services/datasync/ -v`. Expected: PASS.

**Step — commit** `feat(sync): shared datasync service with diff/apply engine`

### Task 0.4: sqlc additions

**Files:**
- Modify: `internal/db/query/*.sql` (roles, items, ability_info, perk_info, category, sync queries)
- Regenerate: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0` then `sqlc generate` (pin v1.30.0 per generated header) — verify `git diff` only touches intended files
- Add: `sync_source` / `sync_run` queries (`ListSyncSources`, `GetSyncSourceByName`, `UpdateSyncSourceURL`, `SetSyncSourceEnabled`, `CreateSyncRun`, `ListSyncRuns`)

**Step — verify:** `go build ./...` + `go vet ./...` clean.

**Step — commit** `feat(db): exact-name lookup + sync queries`

---

## Phase 1 — Sync UI (`/sync`)

### Task 1.1: Sync handler + routes

**Files:**
- Create: `internal/web/handlers/sync.go` (`SyncHandler` with `dbPool` + `*datasync.Service`)
- Modify: `internal/web/server.go` — wire `syncHandler`, add routes:
```go
protected.GET("/sync", syncHandler.Page)
protected.POST("/sync/preview", syncHandler.Preview)      // fetch + diff (rate-limited)
protected.POST("/sync/apply", syncHandler.Apply)          // apply plan (rate-limited)
protected.POST("/sync/sources/:id", syncHandler.UpdateSource) // edit URL/enabled
protected.POST("/sync/run/:id/rerun", syncHandler.RerunSource) // re-preview single source
```
- Apply redeploy-style rate limiting to `/sync/preview` + `/sync/apply` (reuse `redeployRate` constants or add `syncRate`).
- Startup seeding: in `web.New` (or a `datasync.Service` constructor), `INSERT ... ON CONFLICT (name) DO NOTHING` the four sources with URLs from env — so the UI has rows even before first use; the CLI keeps reading env directly.

### Task 1.2: Sync templates

**Files:**
- Create: `internal/web/templates/pages/sync_templ.go` (`.templ` source, commit generated)
- Create: `internal/web/templates/partials/sync_diff_templ.go`, `sync_sources_templ.go`, `sync_run_history_templ.go`

Page layout (mobile-first, obsidian-glass):
- **Callout banner** (top, amber info styling): "These syncs pull the latest roles/items from the Google Sheets. They're normally run **once before a game starts** and not re-run mid-game — preview before applying." (User note, 2026-08-08.)
- **Sources card:** 4 rows (Good/Evil/Neutral Roles, Items) — name, kind, alignment badge, enabled toggle, URL (editable inline via `hx-post`), last run status (latest `sync_run` per source).
- **Preview card:** "Fetch & Preview Changes" button → HTMX-swapped diff table: per-source group headers with counts (🟢 N new / 🟡 M updated / ⚪ S skipped), expandable rows (`<details>`) showing old→new values; a per-source "Apply" button + one "Apply All".
- **History card:** last N `sync_run` rows (time, source, status badge, counts JSON rendered as chips, run_by).
- Toasts on every mutation via existing `HX-Trigger: showToast` pattern.

### Task 1.3: Handler tests

**Files:**
- Create: `tests/web/sync_test.go` — httptest + local DB (bootstrap via `tests/testutil`): login → GET `/sync` 200; POST `/sync/preview` with seeded sources pointing at a local fixture CSV URL (or a `datasync.Service` with injected fetcher — prefer an `io.Reader`-injectable fetch so tests never hit the network) → assert diff table in response; POST `/sync/apply` → assert rows created + `sync_run` row; tokenless POST → 400 (CSRF gate).
Run: `go test ./tests/web/ -v`. Expected: PASS. Then `go test ./...` (full suite, <60s).

**Step — commit** `feat(web): spreadsheet sync page with preview/diff/apply`

---

## Phase 2 — Migrations UI (`/admin/migrations`)

### Task 2.1: Migrations handler + routes

**Files:**
- Create: `internal/web/handlers/migrations.go`
- Modify: `internal/web/server.go`:
```go
protected.GET("/admin/migrations", migrationsHandler.Page)
protected.POST("/admin/migrations/up", migrationsHandler.Up)         // apply pending (rate-limited)
protected.POST("/admin/migrations/down", migrationsHandler.Down)     // rollback N (rate-limited, confirm phrase)
```
- Handler holds `*dbmigrate.Runner` + a `isProd` flag (DSN contains `roundhouse.proxy.rlwy.net` — reuse the existing prod-detection string from `main.go`).

### Task 2.2: Migrations templates

**Files:**
- Create: `internal/web/templates/pages/migrations_templ.go`, `internal/web/templates/partials/migration_table_templ.go`

Page:
- Banner (danger styling) when `isProd`: "Connected to PRODUCTION database."
- Table: version, name, applied_at, dirty badge; current version highlighted.
- Controls: **Apply Pending** (POST `/admin/migrations/up`), **Rollback 1 / Rollback N** (number input, POST `/admin/migrations/down`). When `isProd`, rollback requires typing a confirmation phrase (e.g. the migration name) into a modal/field — server validates before executing. Up against prod is allowed (that's the point of the page) but also gets the banner + a one-click confirm.
- Results render via toast + refreshed table partial.

### Task 2.3: Tests

**Files:**
- Create: `tests/web/migrations_test.go` — httptest against a fresh local DB **at an old version** (bootstrap applies all migrations; instead use a dedicated scratch schema/DB where the runner is pointed at `DATABASE_URL` and the suite migrates down to a known version first): assert table lists applied/pending, POST `/admin/migrations/up` applies pending (version bumps), prod-guard rejects rollback without the confirmation phrase. Also `internal/db/migrate/migrate_test.go` from Task 0.2 covers runner-level up/down.
Run: `go test ./tests/web/ -run Migrations -v` + full suite. Expected: PASS.

**Step — commit** `feat(web): database migrations admin page`

---

## Phase 3 — Finish the admin command port

### Task 3.1: Player-edit gaps (the rest of `/inv`)

**Files:**
- Modify: `internal/web/handlers/player_edit.go` (add handlers) + `internal/web/server.go` (routes)
- Modify: `internal/web/templates/pages/player_edit_templ.go` (form sections)
- Modify: `internal/db/query/*.sql` + regen (exact-name player queries as needed)

Routes (all `POST`, HTMX swap partials, toasts):
```go
protected.POST("/players/:id/coins", playerEditHandler.SetCoins)
protected.POST("/players/:id/luck", playerEditHandler.SetLuck)
protected.POST("/players/:id/item-limit", playerEditHandler.SetItemLimit)
protected.POST("/players/:id/immunity/add", playerEditHandler.AddImmunity)
protected.POST("/players/:id/immunity/remove", playerEditHandler.RemoveImmunity)
protected.POST("/players/:id/death", playerEditHandler.SetDeath)          // alive/dead toggle + kill reason
protected.POST("/players/:id/notes", playerEditHandler.SetNotes)
protected.POST("/players/:id/alignment", playerEditHandler.SetAlignment)
protected.POST("/players/:id/role", playerEditHandler.AssignRole)         // role picker (from DB)
protected.POST("/players", playerEditHandler.Create)                      // /inv create port
```
Semantics mirror `internal/commands/inv/{coin,luck,immunity,death,notes,alignment,role,create}.go` exactly (constants, error strings, item-limit from `game_config` default). Extract shared mutation logic into `internal/services/inventory` if the command handlers need it too — but per the archive policy, the **web handlers are the source of truth**; commands stay as-is.

**Tests:** extend `tests/web/server_test.go` (player edit flows) + `tests/inventory/` additions where logic moved. Run `go test ./...`.

**Commit** `feat(web): complete player admin (coins/luck/immunity/death/notes/alignment/role/create)`

### Task 3.2: Channel-config mutations (the rest of `/channel`)

**Files:**
- Modify: `internal/web/handlers/channels.go` (mutations + channel picker)
- Modify: `internal/web/templates/pages/channels_templ.go`
- Modify: `internal/web/server.go`

Routes:
```go
protected.POST("/channels/vote/set", channelsHandler.SetVoteChannel)
protected.POST("/channels/action/set", channelsHandler.SetActionChannel)
protected.POST("/channels/lifeboard/set", channelsHandler.SetLifeboardChannel)
protected.POST("/channels/admin/add", channelsHandler.AddAdminChannel)
protected.POST("/channels/admin/remove", channelsHandler.RemoveAdminChannel)
protected.POST("/channels/log/set", channelsHandler.SetLogChannel)
protected.POST("/channels/log/remove", channelsHandler.RemoveLogChannel)
```
- UI: each singleton row gains a "Set" control — a `<select>` of guild channels from `discordSession` when Discord is connected (fallback: manual channel-ID text input when `DISABLE_DISCORD=true` or session nil, matching the existing "unverified" status path).
- Deletions/removals use the confirm pattern (`hx-confirm`) — cheap, no extra infra.
- Mirrors `internal/commands/channels/{admin,vote,action,lifeboard,log}.go` semantics (singleton upsert vs multi-row admin list; lifeboard is delete+rebuild+pins — **pin/rebuild requires Discord**, so in web-only mode store the channel and flag it as pending-Discord action; document in the page).

**Tests:** `tests/web/channels_test.go` — set vote channel (DB row upsert), admin add/remove (row create/delete), log set/remove, and CSRF gate. Discord-dependent paths (lifeboard pin) assert DB write + "needs Discord" note only.

**Commit** `feat(web): channel configuration mutations`

### Task 3.3: `/setup` role-pool generator

**Files:**
- Create: `internal/services/roledraft/roledraft.go` (move `activeRoleList`, `generateRoleSelectPool`, `generateRolePools`, `groupRoles` from `internal/commands/setup/setup.go`; keep `ken` out — pure DB + math)
- Modify: `internal/commands/setup/setup.go` (thin wrapper over `roledraft`)
- Create: `internal/web/handlers/setup.go` + `internal/web/templates/pages/setup_templ.go`
- Modify: `internal/web/server.go`:
```go
protected.GET("/setup", setupHandler.Page)
protected.POST("/setup/generate", setupHandler.Generate) // returns partial with pools
```
- Page: player count + deceptionist count form (same validation: ≤ available roles; default deceptionist count = Deceptionist-role members when Discord connected, else manual). Output: deceptionist options cards + two-column random pool, matching the embed layout. Add "copy as text" (`navigator.clipboard`) since the Discord embed is going away as the primary surface.

**Tests:** `internal/services/roledraft/roledraft_test.go` (pure functions: pool sizes, deceptionist caps, alignment grouping). `tests/web/setup_test.go` (form → partial). Run `go test ./...`.

**Commit** `feat(web): game setup role-pool generator`

---

## Phase 4 — Navigation, docs, cleanup

### Task 4.1: Nav additions
- Modify: `internal/web/templates/layouts/base_templ.go` — desktop top nav + mobile bottom nav get: **Sync** (`/sync`), **Setup** (`/setup`), and under an Admin cluster: **Migrations** (`/admin/migrations`). Mobile bottom nav already has 7 items → switch to the scrollable pattern (`flex overflow-x-auto`, `min-w-[76px]`) per the go-web-admin-panels skill.
- `make generate` and commit `_templ.go` + `output.css`.

### Task 4.2: Docs + archive policy
- Modify: `AGENTS.md` — command inventory table gains a "Web status" column; add "Admin commands are archived (2026-08): web panel is the admin surface; do not maintain parity in slash commands" note; document `/sync`, `/admin/migrations`, `/setup`, new env vars (`GOOD_ROLES_CSV` etc. → now seeded into `sync_source`); update migration paths (embedded dir), Makefile targets.
- Modify: `internal/web/README.md` — full route table refresh incl. new routes; roadmap strikethroughs for completed items.
- `.env.example`: document the four CSV vars.

### Task 4.3: Final verification
1. `make db-up` (once) → `go test ./...` → all green.
2. `make build` → binary builds; `make run-web` boots with local DB (`DATABASE_POOLER_URL=$(grep '^DATABASE_URL=' .env | cut -d= -f2-) make run-web` — never iterate against prod).
3. Browser pass: `/sync` preview/apply against local DB with a fixture CSV; `/admin/migrations` up; `/setup` generate; player edit + channels mutations.
4. `git status` clean of binaries; PR per phase with the 3-round review workflow (spec compliance → code quality → independent review).

---

## Files Likely to Change (summary)

| Area | Files |
|---|---|
| New services | `internal/services/datasync/{parse,plan,apply,sources}.go`, `internal/db/migrate/migrate.go`, `internal/services/roledraft/roledraft.go` |
| Migrations | `internal/db/migration/000030_*.sql`, `000031_*.sql`; **moved** to `internal/db/migrate/migrations/` |
| sqlc | `internal/db/query/*.sql` + regenerated `internal/models/*.sql.go` |
| Web handlers | new `sync.go`, `migrations.go`, `setup.go`; extended `player_edit.go`, `channels.go` |
| Web templates | new `pages/{sync,migrations,setup}_templ.go`, partials; extended `player_edit`, `channels`, `layouts/base` |
| Wiring | `internal/web/server.go`, `cmd/betrayal-bot/main.go`, `internal/web/railway` (n/a), `Makefile` |
| CLI | `cmd/data-entry/main.go` → thin wrapper |
| Tests | `internal/services/datasync/*_test.go`, `internal/db/migrate/migrate_test.go`, `internal/services/roledraft/roledraft_test.go`, `tests/web/{sync,migrations,channels,setup}_test.go`, `tests/testutil` path fix |
| Docs | `AGENTS.md`, `internal/web/README.md`, `.env.example` |

## Verification / Acceptance

- `go vet ./...`, `go build ./...`, `go test ./...` all green at every phase boundary (headless — no Discord, no game window).
- Sync: preview shows exact old→new diffs; apply is transactional with `sync_run` audit; CLI still works via env vars.
- Migrations: table reflects real `schema_migrations` state; up/down work; prod guard blocks rollback without confirmation.
- Admin parity: every row in the Part 1a table marked ✅ has a working web route; Discord admin commands untouched (archive).

## Risks, Tradeoffs, Open Questions

**Resolved (2026-08-08, user-approved):**
1. **Production target** — destructive web actions (`/sync/apply`, `/admin/migrations/down`, `/admin/migrations/up`) are intentionally available because this panel is the production operations surface. The UI clearly labels the live database and the sync page carries a preview-first callout.
2. **Sync semantics** — upsert: sheet edits propagate to existing rows (that's the point of preview/validate/apply).
3. **`cmd/data-entry`** — kept as a thin CLI wrapper (archive mode).
4. **Admin `/roll`** — NOT ported; rolls stay Discord-only for now.
5. **Lifeboard** — web sets the stored channel; the Discord-side rebuild+pin stays a bot operation (page flags it when Discord is disabled).
6. **Migrations embedded** — SQL files move to `internal/db/migrate/migrations/` (single source of truth); Makefile + testutil updated.

**Remaining watch items:**
- **Prod-pooler hazard** — `make run-web` hits prod. The production banner is intentionally visible; operators must preview sync changes and use the server-side confirmations for destructive actions.
- **Migration move** touches Makefile/testutil paths — do it in the same commit as the embed so nothing is ever broken in between.
