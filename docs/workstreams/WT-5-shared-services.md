# WT-5 — Shared Application Services

> **Historical / superseded:** This workstream predates the completed SvelteKit migration. References to legacy web handlers are retained only as historical implementation context.

## Branch/worktree

- Branch: `wt-shared-services`
- Worktree: `../betrayal-shared-services`
- Migration ownership: none by default; request a slot before adding one

## Mission

Make Discord commands and web admin routes call the same application services for shared game behavior. This is incremental extraction, not a rewrite.

## Scope

Owned paths:

- `internal/services/`
- Selected `internal/commands/` call sites
- Selected `internal/web/handlers/` call sites
- Focused unit/integration tests for extracted services

Do not modify startup/environment files, CI, Dockerfile, generated sqlc code, or unrelated UI templates.

## Required extraction order

1. Inventory/player mutations after WT-1 semantics are established.
2. Notes.
3. Channel configuration.
4. Cycle transitions.
5. Catalog mutations only if the service boundary is clear.

Each extraction must preserve behavior through a vertical test slice:

1. Write a test expressing the shared behavior.
2. Run it and observe the expected failure.
3. Implement the service API minimally.
4. Rewire exactly one caller.
5. Run focused tests.
6. Rewire the second caller.
7. Run focused tests again.
8. Refactor only while green.

## Architecture constraints

- Services accept `context.Context`.
- Services do not import Discord or Echo.
- Authorization policy is explicit, testable, and not inferred from route location.
- Compound mutations use one transaction where atomicity matters.
- Return errors; do not silently convert query errors to empty results.
- Repository/sqlc queries remain the persistence mechanism; do not hand-edit generated files.

## Verification

```bash
go test -count=1 ./internal/services/... ./tests/inventory ./tests/web
go test -race -count=1 ./internal/services/... ./tests/inventory ./tests/web
go vet ./...
```

Then run the complete global gate and verify both Discord-adapter and web-adapter call sites are actually using the service. A service that exists but is unused is not completion.

Commit locally only. Do not push or merge.
