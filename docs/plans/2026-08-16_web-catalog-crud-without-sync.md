# Web catalog CRUD without spreadsheet sync

> **Status:** Complete as of 2026-08-16: phases 1–4 and the whisper verification
> (phase 5) are implemented, tested, and committed on `main` (not pushed).
> Coordinates with the WT-6 web-panel workstream; supersedes the stale
> "web panel lacks item/ability/status CRUD" note in `AGENTS.md` (that CRUD
> landed — see `internal/web/api/catalog.go`).

## Goal

Let hosts add and remove game data from the web admin panel directly, so everyday
catalog/roster changes no longer require editing a spreadsheet and running a full
sync. Sync remains for bulk import; the panel becomes the source for surgical edits.

## Current state (gap matrix)

| Data | Web create/delete today | Notes |
|---|---|---|
| Roles | ✅ (`/roles`, `/api/v1/catalog/roles`) | full CRUD + search |
| Items | ✅ (`/items`) | full CRUD + search |
| Abilities | ✅ (`/abilities`) | full CRUD + search |
| Statuses | ✅ (`/statuses`) | full CRUD + search |
| Players | create ✅ / delete ❌ | `DeletePlayer` query exists; only reachable today via `/inv delete` (Discord) and as an error-rollback in `/inv create` |
| Perks (`perk_info`) | ❌ | sync-only; sqlc CRUD queries already exist (`perk_info.sql`) |
| Categories (`category` + `item_category`/`ability_category`) | ❌ | sync-only; partial queries exist |
| Role ↔ ability/perk links (`role_ability`, `role_perk`) | ❌ | sync-only; join queries exist |

All player-dependent rows use `ON DELETE CASCADE` (`player_item` 000012, `player_status`
000013, `player_perk` 000014, `player_ability` 000015, `player_confessional` 000016,
`player_immunity` 000017, `player_note` 000022, `vote` 000027, `whisper_group_member`
000035). `player_lifeboard` (000021) has **no** FK to `player` — no constraint blocks
deletion; the lifeboard must be rebuilt afterwards via `/channel lifeboard set`.

## Scope (phases)

1. **Player removal** — `DELETE /api/v1/players/:id` (204, 404; auth-gated; cascade
   wipes inventory/confessional/votes/whisper membership) + remove control on the
   player profile with an exact-label typed confirmation. Parity model: `Inv.delete`
   in `internal/commands/inv/create.go` (pinned-message cleanup + `DeletePlayer`).
   Lifeboard: document rebuild; optional best-effort Discord pinned-message cleanup
   when a Discord session is present.
2. **Perk library** — extend `/api/v1/catalog` with `kind=perks` (list/get/create/
   update/delete via existing `perk_info.sql` queries) and a `/perks` SvelteKit page
   reusing `CatalogPage.svelte` (needs a perk input shape in `types.ts`).
3. **Category libraries** — `kind=categories` CRUD on `/catalog` (reuse
   `ListCategory`/`CreateCategory`/`DeleteCategory`; add `UpdateCategory` query) and
   item/ability category assignment: add `ListItemCategoryNames` +
   `DeleteItemCategoryJoin` + `DeleteAbilityCategoryJoin` queries, surface current
   categories in item/ability detail DTOs, and manage them on the detail UI.
4. **Role link editors** — ability + perk multi-selects on the role detail page using
   existing `CreateRoleAbilityJoin` / `DeleteRoleAbilityJoin` /
   `CreateRolePerkJoin` / `DeleteRolePerkJoin`; role DTO already returns linked
   abilities/perks.
5. **Whisper verification (carry-over)** — statistical + pool-randomness regression
   tests for the 5% doubt swap (`SuspicionChance`), and add the `whisper_*` tables to
   `tests/testutil/testutil.go` `allTables` (isolation gap since migration 000035).

## Boundaries

- No new migrations unless a query is truly missing (`UpdateCategory`, join-delete
  queries are additive SQL in `internal/db/query/` + sqlc regen — no schema change).
- No changes to Discord command behavior; `/inv delete` remains the Discord path.
- Player deletion requires exact-label typed confirmation; server rejects the request
  without a valid `confirm` field. No raw browser `confirm()`-only guards.
- Production remains the intended target — no `WEB_ALLOW_PROD_MUTATIONS`-style escape
  hatch; behave like the existing catalog/player mutation routes.
- Do not render new raw snowflakes; reuse the Discord-resources label resolution.

## TDD and verification

- Per phase: write the focused httptest/component test first, run RED, implement the
  smallest change, run GREEN.
- Player delete tests: 401 unauthenticated, 204 + cascade (confessional/inventory
  rows gone), second delete 404, missing player 404.
- Frontend: Vitest component tests for remove control (wrong phrase disables submit,
  cancel sends nothing, confirm sends DELETE and navigates), catalog pages for perks
  and categories, role link editor.
- Gates per phase: focused tests, `go vet ./...`, `go test -count=1 ./...` (local
  Postgres via `make db-up`), `npm --prefix frontend run test:unit` + `check`,
  `make build` (re-generates embedded `internal/web/ui/dist`), then inspect
  `git status`/`git diff` for artifact churn.
- Commit locally only; no push or merge without explicit approval.

## Collision map

- Owns: `internal/web/api/catalog.go`, `internal/web/api/players_admin.go`,
  `internal/web/server.go` (route table), `frontend/src/routes/**`,
  `frontend/src/lib/**`, `internal/db/query/{category,item_category,ability_category}.sql`
  (+ regenerated `internal/models`), `tests/web/**`, `tests/testutil/testutil.go`,
  `docs/plans/`.
- Current checkout (`main`) carries uncommitted whisper UX work (sender status in the
  ephemeral embed, `internal/commands/whisper/whisper.go`,
  `internal/services/whisper/*`). Phase 5's whisper tests build on that in-progress
  state; everything else is independent of it.