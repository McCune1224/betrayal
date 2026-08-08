# WT-2 — Dependency and Container Hardening

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

Commit locally only. Do not push or merge.
