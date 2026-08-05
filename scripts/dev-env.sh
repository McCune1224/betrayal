#!/usr/bin/env bash
# dev-env.sh — worktree & environment tooling for the Betrayal bot.
#
# Problem: .env is gitignored, so every `git worktree add` lands with no
# config. Solution: secrets live in ONE canonical file OUTSIDE the repo
# ($BETRAYAL_ENV_FILE, default ~/.config/betrayal/env) and every checkout's
# .env is a symlink to it.
#
# Commands:
#   link               symlink this checkout's .env -> canonical env file
#   new-worktree NAME  create ../betrayal-NAME on branch wt-NAME, link env,
#                      boot local DB + run migrations
#   doctor             verify env file, required keys, local DB reachability

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CANONICAL_ENV="${BETRAYAL_ENV_FILE:-$HOME/.config/betrayal/env}"

fatal() { echo "!! $*" >&2; exit 1; }
info()  { echo "== $*"; }

# Keys that MUST exist in the canonical env file (from .env.example)
required_keys() {
  grep -E '^[A-Z_]+=' "$REPO_ROOT/.env.example" | cut -d= -f1
}

cmd_link() {
  if [[ ! -f "$CANONICAL_ENV" ]]; then
    info "canonical env not found at $CANONICAL_ENV; scaffolding from .env.example"
    mkdir -p "$(dirname "$CANONICAL_ENV")"
    cp "$REPO_ROOT/.env.example" "$CANONICAL_ENV"
    chmod 600 "$CANONICAL_ENV"
    fatal "edit $CANONICAL_ENV with real values, then re-run: $0 link"
  fi

  if [[ -e "$REPO_ROOT/.env" && ! -L "$REPO_ROOT/.env" ]]; then
    info "$REPO_ROOT/.env is a real file; leaving it untouched (override mode)"
    info "remove it and re-run '$0 link' to switch to the shared env"
    return 0
  fi

  ln -sfn "$CANONICAL_ENV" "$REPO_ROOT/.env"
  info "linked $REPO_ROOT/.env -> $CANONICAL_ENV"
}

cmd_new_worktree() {
  local name="${1:-}"
  local branch="${2:-wt-$name}"
  local dest
  dest="$(dirname "$REPO_ROOT")/betrayal-$name"

  [[ -n "$name" ]] || fatal "usage: $0 new-worktree NAME [branch]"
  [[ -d "$dest" ]] && fatal "$dest already exists"
  [[ "$name" =~ ^[a-z0-9-]+$ ]] || fatal "NAME must be lowercase alnum/hyphen"

  git -C "$REPO_ROOT" worktree add "$dest" -b "$branch"
  info "worktree created: $dest (branch $branch)"

  (cd "$dest" && ./scripts/dev-env.sh link)

  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    (cd "$dest" && make db-up) || info "db-up failed — start postgres manually, then: make mock-migrate-up"
  fi
  (cd "$dest" && make mock-migrate-up) || info "mock-migrate-up failed — check MOCK_DATABASE in the env file"
  info "done. cd $dest to start working (or open it as a Hermes desktop project)"
}

cmd_doctor() {
  [[ -e "$REPO_ROOT/.env" ]] || fatal "no .env in $REPO_ROOT — run '$0 link'"
  if [[ -L "$REPO_ROOT/.env" ]]; then
    info ".env is a symlink -> $(readlink "$REPO_ROOT/.env")"
  else
    info ".env is a real file (override mode)"
  fi

  local missing=0
  for k in $(required_keys); do
    if ! grep -q "^$k=" "$REPO_ROOT/.env" 2>/dev/null; then
      echo "   missing key: $k"
      missing=1
    fi
  done
  [[ $missing -eq 0 ]] || fatal "env file is missing required keys (see .env.example)"

  local mock
  mock="$(grep -E '^MOCK_DATABASE=' "$REPO_ROOT/.env" | cut -d= -f2- | tr -d '"')"
  if [[ -n "$mock" ]]; then
    info "MOCK_DATABASE: ${mock%%\?*}"
  fi
  info "doctor OK — env file looks complete"
}

cmd="${1:-}"
case "$cmd" in
  link)         cmd_link ;;
  new-worktree) cmd_new_worktree "${2:-}" "${3:-}" ;;
  doctor)       cmd_doctor ;;
  *)
    echo "usage: $0 {link|new-worktree NAME [branch]|doctor}"
    exit 1
    ;;
esac
