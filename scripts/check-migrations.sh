#!/usr/bin/env bash
set -euo pipefail

# Every migration version must have exactly one up and one down file. This
# catches the production-fatal class where a database has advanced to a
# version the release cannot read or roll back.
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dir="$root/internal/db/migrate/migrations"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

for file in "$dir"/*.sql; do
  base="${file##*/}"
  if [[ "$base" =~ ^([0-9]+)_.+\.(up|down)\.sql$ ]]; then
    printf '%s %s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" >>"$tmp"
  fi
done

fail=0
while read -r version; do
  up_count=$(awk -v v="$version" '$1 == v && $2 == "up" { count++ } END { print count + 0 }' "$tmp")
  down_count=$(awk -v v="$version" '$1 == v && $2 == "down" { count++ } END { print count + 0 }' "$tmp")
  if [[ "$up_count" -ne 1 || "$down_count" -ne 1 ]]; then
    printf 'migration %s must have exactly one up and one down file (found up=%s down=%s)\n' "$version" "$up_count" "$down_count" >&2
    fail=1
  fi
done < <(awk '{ print $1 }' "$tmp" | sort -u)

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

echo "migration file pairs valid: $(wc -l < "$tmp") files"
