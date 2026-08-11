# Guild-aware web configuration and production sync sources

## Goal

Make the admin panel use the configured Discord guild as a source of truth for channel selectors and make production CSV-source configuration observable and reliable.

## Worktree

- Branch: `wt-guild-aware-web-sync`
- Worktree: `/home/mckusa/Code/betrayal-guild-aware-web-sync`
- Base: clean `main` at the coordinator commit

## Scope

1. Add explicit Discord guild configuration to startup/web config with `DISCORD_GUILD_ID` fallback to the existing constant.
2. Add a small Discord metadata provider for the configured guild's channels, with display name, category, type, and ID.
3. Change `/channels` to render known guild channels as dropdown options for vote/action/lifeboard/admin/log updates. Preserve a raw-ID fallback only when Discord is unavailable.
4. Surface guild connection state and channel-list failures in the page instead of silently falling back to IDs.
5. Make CSV source seeding report whether each URL came from the environment, remains blank, or was manually configured; keep manually edited URLs intact.
6. Add startup logging with names and presence/absence only—never URL values—and tests proving env-key mapping and seed behavior.

## Boundaries

- Do not run or alter production migrations.
- Do not change Discord command behavior.
- Do not expose CSV URL values in logs.
- Do not overwrite manually edited `sync_source.url` values from environment variables.
- No changes to local env secrets.
- Generated templ files may be regenerated only for source template changes.

## TDD and verification

- First add failing tests for guild channel option presentation and CSV env-source visibility.
- Run focused tests RED before implementation.
- Implement the smallest root-cause changes, then rerun GREEN.
- Verify with focused tests, `go test -race -p 1 ./...`, `go vet ./...`, `go mod tidy -diff`, `make build`, and `git diff --check`.
- Inspect generated CSS/status after build and restore verification-only artifacts.
- Commit locally only; do not push or merge.

## Collision map

- This stream owns `internal/web/`, `internal/services/datasync/`, startup wiring in `cmd/betrayal-bot/main.go`, generated templ files, and related tests.
- It must sequence after or be merged carefully with any branch editing `Makefile`, CI, migrations, or shared startup config.
- The existing `wt-twin` checkout is dirty and remains isolated; this branch starts from clean `main` and does not include its uncommitted twin feature.
