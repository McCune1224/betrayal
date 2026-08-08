#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf 'dependency/container validation failed: %s\n' "$1" >&2
  exit 1
}

image=''
if [[ "${1:-}" == '--image' ]]; then
  [[ $# -eq 2 ]] || fail '--image requires exactly one image name'
  image=$2
elif [[ $# -ne 0 ]]; then
  fail "unknown argument: $1"
fi

[[ -f Dockerfile ]] || fail 'Dockerfile is missing'
[[ -f .dockerignore ]] || fail '.dockerignore is missing'

if grep -nE '(^|[[:space:]])(ARG|ENV)[[:space:]]+ENV_FILE|COPY[[:space:]].*\.env' Dockerfile; then
  fail 'Dockerfile must not bake an environment file into an image'
fi

for pattern in '**/.env' '.git' 'bin' '**/*secret*' '**/*token*' '**/*password*' '**/*.pem' '**/*.key'; do
  grep -Fqx "$pattern" .dockerignore || fail ".dockerignore must exclude $pattern"
done

grep -qE '^go 1\.25(\.0)?$' go.mod || fail 'pgx/v5 fixed release requires the Go 1.25 toolchain'
grep -qE 'github\.com/jackc/pgx/v5 v5\.(9\.[2-9]|9\.[0-9]{2,}|10\.)' go.mod || fail 'pgx/v5 must be upgraded to at least v5.9.2'

mapfile -t changed_go_files < <(
  {
    git diff --name-only HEAD^..HEAD -- '*.go' 2>/dev/null || true
    git diff --name-only -- '*.go'
    git diff --cached --name-only -- '*.go'
  } | sort -u
)
if ((${#changed_go_files[@]} > 0)); then
  formatted=$(gofmt -l -- "${changed_go_files[@]}")
  [[ -z "$formatted" ]] || fail "gofmt reports changed Go files:\\n$formatted"
fi

go mod tidy -diff

if [[ -n "$image" ]]; then
  history=$(docker history --no-trunc "$image") || fail "unable to inspect Docker history for $image"
  if printf '%s\n' "$history" | grep -Eiq 'ENV_FILE|\.env([[:space:]]|$)|([[:space:]]|^)(secret|token|password|credential|api[_-]?key)([=:[:space:]]|$)'; then
    printf '%s\n' "$history" >&2
    fail 'Docker history contains secret-bearing arguments, environment, or files'
  fi

  container=$(docker create "$image") || fail "unable to create container from $image"
  cleanup() { docker rm "$container" >/dev/null; }
  trap cleanup EXIT
  contents=$(docker export "$container" | tar -tf -) || fail 'unable to inspect image contents'
  if printf '%s\n' "$contents" | grep -Eiq '(^|/)(\.env|[^/]*(secret|token|password|credential|api[_-]?key)[^/]*)$|\.(pem|key|p12|pfx)$'; then
    printf '%s\n' "$contents" >&2
    fail 'image contains secret-like files'
  fi
fi

printf 'dependency/container validation passed\n'
