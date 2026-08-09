# WT-3 — Environment and Operations Safety

## Branch/worktree

- Branch: `wt-environment-operations`
- Worktree: `../betrayal-environment-operations`
- Migration ownership: none unless explicitly assigned
- Shared-file ownership: `cmd/betrayal-bot/main.go`, `Makefile`, `.env.example`, active operations docs, `scripts/dev-env.sh`

## Mission

Make local development structurally unable to default to the production database and make production migrations explicit.

## Scope

Owned paths:

- `cmd/betrayal-bot/main.go`
- `Makefile`
- `.env.example`
- `scripts/dev-env.sh`
- `AGENTS.md`, `README.md`, `docs/operations.md`, `internal/web/README.md` only for environment/operations documentation
- Tests for startup configuration and Makefile/script safety

Do not modify inventory/domain logic, Dockerfile, CI workflow, or Discord command behavior except for configuration plumbing.

## Required behavior tests first

1. `ENVIRONMENT=local` selects `DATABASE_URL`.
2. `ENVIRONMENT=production` selects `DATABASE_POOLER_URL`.
3. Local startup fails if `DATABASE_URL` is absent rather than silently falling back to the pooler.
4. Production startup fails if the production DSN is absent.
5. Production migration commands require explicit confirmation.
6. Test bootstrap remains local-only and strips/ignores production DSNs.

8. Worktree doctor validates the environment contract without printing secret values.

Run each focused test red before implementing.

## Design requirements

- Extract environment loading/validation into testable functions; do not leave all logic in `main()`.
- Keep production detection conservative and explicit.
- Preserve the current local DB safety guard.
- Make `make run-web` safe by default for local development.
- Avoid shell parsing that truncates values containing `=`.
- Never echo secret values.

## Verification

```bash
go test -count=1 ./cmd/betrayal-bot ./tests/testutil ./internal/web/...
go test -race -count=1 ./cmd/betrayal-bot ./tests/testutil ./internal/web/...
./scripts/dev-env.sh doctor
make -n run-web
make -n migrate-local-up
```

Then run the complete global gate. Inspect `git diff` for accidental secret/path disclosure.

Commit locally only. Do not push or merge.
