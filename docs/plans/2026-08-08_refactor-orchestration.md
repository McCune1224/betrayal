# Betrayal Refactoring Program

> **Completed / historical:** This program plan covers the pre-SvelteKit refactor sequence. The SvelteKit migration and templ/HTMX cutover are complete; use the root `AGENTS.md` and [`internal/web/README.md`](../../internal/web/README.md) for current guidance.

**Status:** Completed
**Owner:** Project manager / main branch
**Started:** 2026-08-08
**Strategy:** TDD-first, isolated worktrees, verified merge gates

## Objective

Reduce correctness, security, operational, and architectural risk in the multi-year Betrayal bot without performing a broad rewrite. Every behavior change must be driven by a test that was observed failing before production code was changed.

## Non-negotiable engineering rules

1. **No production change without a failing test first.**
2. Each test must fail for the intended behavioral reason, not because of a typo or broken setup.
3. Use vertical RED -> GREEN -> REFACTOR slices. Do not write a large test pile and implement later.
4. Tests must exercise real behavior. Avoid mocks unless the external dependency is genuinely unavoidable.
5. Every bug fix must include a regression test.
6. Every touched package must pass focused tests, then the complete suite.
7. No worktree may use production data for tests or manual mutation checks.
8. No agent may modify files outside its assigned ownership without coordinating with the project manager.
9. Generated files are regenerated with pinned tools and are never hand-edited.
10. Agents commit locally only. No pushing, merging, rebasing, or deleting branches without project-manager direction.

## Global verification gate

Before a workstream can be marked complete:

```bash
gofmt -l .                         # must print nothing
go vet ./...
go test -race ./...
go test -count=1 ./...
go mod tidy -diff                 # must exit cleanly
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
make build                         # required when build/template/CSS paths are touched
git diff --check
git status --short --branch
```

If the vulnerability command is unavailable due to network/toolchain issues, report the exact blocker; do not substitute a claim that the scan passed.

## Workstream dependency order

### Foundation / safety first
- **WT-1 inventory-correctness**: authorization and inventory/note correctness. Lowest collision risk; immediately reduces player-state risk.
- **WT-2 dependency-container**: vulnerable dependencies, Docker secret handling, gofmt/tidy gates. Shared build files owned here.
- **WT-3 environment-operations**: local/prod DSN selection and safe migration commands. Owns `main.go`, Makefile, `.env.example`, operations docs.

### Then architectural extraction
- **WT-4 discord-boundary**: Discord context/guild/config/audit boundaries. Must follow WT-3's environment contract; owns `internal/discord`, selected command adapters, and Discord-specific tests.
- **WT-5 shared-services**: move shared rules behind application services used by both Discord and web. Must follow WT-1 correctness semantics; owns service interfaces and selected web/command call sites.

### Sequencing rules

- WT-2 owns CI workflow changes. Other workstreams may propose CI checks in their brief but must not edit `.github/workflows/test.yml`.
- WT-3 owns `cmd/betrayal-bot/main.go`, Makefile, `.env.example`, and active operations docs.
- WT-4 may modify `main.go` only through a coordinated handoff with WT-3.
- WT-5 may modify web/command call sites but must not rewrite shared startup or CI files.
- Only one branch may add migrations at a time. WT-1 and WT-5 should avoid schema changes; if a migration is unavoidable, request a numbered slot from the project manager before writing it.

## Merge protocol

1. Inspect every worktree: branch, status, commits, and diff stat.
2. Verify the committed branch independently; do not trust an agent's summary.
3. Read security-sensitive diffs manually: DSN selection, auth, production guards, Docker, migrations, Discord permissions/intents.
4. Merge in dependency order: WT-1 -> WT-2 -> WT-3 -> WT-4 -> WT-5.
5. After every merge, run the complete global verification gate.
6. Regenerate templ/CSS after conflict resolution, then inspect generated diffs.
7. Never remove a worktree until its branch is merged or explicitly abandoned and its status is clean.
8. Keep a merge ledger in this document with commit IDs and verification output.

## Worktree environment

- Canonical env: `~/.config/betrayal/env`, mode 600.
- Each worktree `.env` should be a symlink to the canonical file unless it intentionally has a real per-worktree override.
- Tests must use local `DATABASE_URL`; never `DATABASE_POOLER_URL`.
- The shared Docker Postgres is started once per machine. Other worktrees reuse it.
- Run `./scripts/dev-env.sh doctor` in every worktree before tests.

## Definition of done

A workstream is not done because code compiles. It is done only when:

- The behavior has a test-first history in the commit/agent report.
- Focused tests pass.
- Full `go test -race ./...` passes.
- `go vet ./...` passes.
- `gofmt -l .` is empty.
- No production/test DB safety rule was weakened.
- No secrets or binaries are tracked.
- Documentation reflects the actual behavior.
- The branch is committed and the worktree is clean.

## Workstream briefs

- `docs/workstreams/WT-1-inventory-correctness.md`
- `docs/workstreams/WT-2-dependency-container.md`
- `docs/workstreams/WT-3-environment-operations.md`
- `docs/workstreams/WT-4-discord-boundary.md`
- `docs/workstreams/WT-5-shared-services.md`
