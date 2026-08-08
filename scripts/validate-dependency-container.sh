#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf 'dependency/container validation failed: %s\n' "$1" >&2
  exit 1
}

if grep -nE '(^|[[:space:]])(ARG|ENV)[[:space:]]+ENV_FILE|COPY[[:space:]].*\.env' Dockerfile; then
  fail 'Dockerfile must not bake an environment file into an image'
fi

grep -qE '^\*\*/\.env$' .dockerignore || fail '.dockerignore must exclude .env files'
grep -qE 'github\.com/jackc/pgx/v5 v5\.(9\.[2-9]|9\.[0-9]{2,}|10\.)' go.mod || fail 'pgx/v5 must be upgraded to at least v5.9.2'
grep -qE 'golang\.org/x/text v0\.(39|[4-9][0-9])\.' go.mod || fail 'golang.org/x/text must be upgraded to at least v0.39.0'

changed_go_files=$(git diff --name-only HEAD -- '*.go')
if [[ -n "$changed_go_files" ]] && gofmt -l $changed_go_files | grep -q .; then
  fail 'gofmt reports changed Go files'
fi

go mod tidy -diff
printf 'dependency/container validation passed\n'
