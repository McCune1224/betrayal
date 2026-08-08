# WT-1 — Inventory Correctness

## Branch/worktree

- Branch: `wt-inventory-correctness`
- Worktree: `../betrayal-inventory-correctness`
- Migration ownership: none; do not add migrations

## Mission

Fix the confirmed inventory authorization and quantity/note correctness bugs from the 2026-08-08 systems audit. Treat this as a TDD bug-fix stream, not a redesign.

## Scope

Owned paths:

- `internal/services/inventory/`
- `internal/commands/inv/` only when required to expose/fix service behavior
- `tests/inventory/`
- Focused tests under `internal/services/inventory/`

Allowed generated/query changes: none unless the project manager explicitly approves them.

## Required RED/GREEN cases

Write each test first and run it to observe the expected failure:

1. A non-admin cannot mutate another player's inventory from that player's confessional.
2. The player can operate on their own inventory in their own confessional.
3. An admin can operate in a player's confessional.
4. A non-admin cannot use an admin-whitelisted channel to mutate inventory.
5. Removing quantity 1 decrements by 1.
6. Removing quantity N decrements by N.
7. Removing more than owned deletes the join or clamps according to the established command contract; document the chosen behavior in the test.
8. Updating note position 1 changes note 1, not the last note.
9. Deleting note position 1 deletes note 1, not the last note.
10. `UpdateAbility` returns not-found when the player does not possess the ability.
11. Database/query errors are returned instead of converted into zero values.

## Implementation constraints

- Do not change product semantics beyond the bug under test.
- Pass bounded contexts into every touched DB operation.
- Do not use `log.Println`; use the project logger or return errors.
- Do not make web-only behavior different from Discord behavior.
- Do not refactor unrelated command packages.

## Verification

Focused:

```bash
go test -count=1 ./tests/inventory ./internal/services/inventory
go test -race -count=1 ./tests/inventory ./internal/services/inventory
```

Then run the complete gate from `docs/plans/2026-08-08_refactor-orchestration.md`.

## Handoff report

Report:

- Every RED test command and the expected failure
- Implementation commit(s)
- Focused and full verification output
- Any behavior decision made for over-removal/negative quantities
- Files changed and files intentionally not changed

Commit locally only. Do not push or merge.
