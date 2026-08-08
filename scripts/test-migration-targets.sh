#!/usr/bin/env bash
# Validate migration target compatibility and production safety without running migrations.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

declare -A production_targets=(
  [migrate-up]=migrate-production-up
  [migrate-down]=migrate-production-down
  [migrate-sync]=migrate-production-sync
)

for target in "${!production_targets[@]}"; do
  production_target="${production_targets[$target]}"
  dry_run="$(make -n "$target")"
  production_dry_run="$(make -n "$production_target")"
  [[ "$dry_run" == "$production_dry_run" ]] || {
    echo "migration target $target does not delegate to $production_target" >&2
    exit 1
  }
  grep -Fq 'CONFIRM_PRODUCTION_MIGRATION' <<<"$dry_run" || {
    echo "migration target $target is missing the production confirmation gate" >&2
    exit 1
  }
done

for target in "${!production_targets[@]}"; do
  if make "$target" >/tmp/betrayal-migrate-targets.out 2>/tmp/betrayal-migrate-targets.err; then
    echo "$target unexpectedly ran without explicit production confirmation" >&2
    exit 1
  fi
  grep -Fq 'refusing production migration' /tmp/betrayal-migrate-targets.err || {
    echo "$target refusal did not identify the missing confirmation" >&2
    cat /tmp/betrayal-migrate-targets.err >&2
    exit 1
  }
done

echo "migration target compatibility and production refusal checks passed"
