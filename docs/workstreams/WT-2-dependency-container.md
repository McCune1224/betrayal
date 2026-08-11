# WT-2 — Dependency and Container Hardening

> **Historical / superseded:** This workstream predates the completed SvelteKit migration. Its templ/Tailwind dependency guidance is no longer applicable; use the current `go.mod`, `Dockerfile`, and `AGENTS.md`.

## Branch/worktree

- Branch: `wt-dependency-container`
- Worktree: `../betrayal-dependency-container`
- Migration ownership: none
- Shared-file ownership: `go.mod`, `go.sum`, `Dockerfile`, `.dockerignore`, `.github/workflows/test.yml`, gofmt/tidy/vulnerability CI checks

## Mission

Remove reachable dependency vulnerabilities and make container builds incapable of baking application secrets into image layers.

## Scope

Owned paths:

- `go.mod`, `go.sum`
- `Dockerfile`
- `.dockerignore`
- `.github/workflows/test.yml`
- `Makefile` only if needed for dependency/build verification; coordinate before touching it
- Documentation specifically describing container/dependency verification

Do not modify application behavior, `cmd/betrayal-bot/main.go`, service code, migrations, or Discord command logic.

## Required tests/verification

Dependency work must be verified as a real upgrade, not just a version edit:

```bash
go mod tidy -diff
go vet ./...
go test -race ./...
go test -count=1 ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Container work must verify:

```bash
docker build --build-arg ENV_FILE=.env.example -t betrayal:local-build-test .
docker history betrayal:local-build-test
```

The final Dockerfile must not require `ENV_FILE` and must not copy `.env` into the image. Use a secret-free build context and runtime env injection.

## TDD rule

Configuration/container changes may use executable CI or shell validation tests instead of Go tests, but each new invariant must have a failing validation before implementation. Examples:

- A test/script that fails when `Dockerfile` contains `COPY ... .env`
- A test/script that fails when `go mod tidy -diff` is dirty
- A test/script that fails when gofmt reports files

## Constraints

- Keep pinned templ/Tailwind versions unless compatibility is explicitly tested.
- Upgrade pgx to a version fixing the reported reachable vulnerability.
- Do not blindly upgrade every dependency.
- Do not commit binaries, `.env`, or generated secret material.
- Report any vulnerability that remains and why.

## Toolchain compatibility evidence

The reachable pgx SQL-injection fix is `github.com/jackc/pgx/v5 v5.9.2`.
The first fixed release requires Go 1.25 (`go` directive in its module file);
the last Go-1.23-compatible release tested here, v5.7.6, still reports
GO-2026-5004 in `internal/models/command_audit.sql.go`. Therefore this
workstream intentionally raises the module, Docker image, and CI setup-go
toolchain to Go 1.25 rather than retaining the repository's previous Go 1.23
guidance. Pinned templ/Tailwind versions remain unchanged.

Commit locally only. Do not push or merge.
