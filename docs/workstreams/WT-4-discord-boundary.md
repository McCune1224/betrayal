# WT-4 — Discord Boundary and API Hygiene

## Branch/worktree

- Branch: `wt-discord-boundary`
- Worktree: `../betrayal-discord-boundary`
- Migration ownership: none
- Shared-file constraint: do not edit `cmd/betrayal-bot/main.go` without coordination with WT-3

## Mission

Reduce Discord-specific coupling, eliminate hardcoded interaction assumptions, make guild/config behavior explicit, and make command audit logging accurate and defensive.

## Scope

Owned paths:

- `internal/discord/`
- Discord adapter portions of `internal/commands/` only when needed
- `internal/logger/audit_commands.go`
- `cmd/betrayal-bot/main_test.go` and new Discord-focused tests
- Configuration plumbing only through a coordinated WT-3 handoff

Do not modify inventory service semantics, web handlers, Dockerfile, Makefile, CI, or migrations.

## Required TDD slices

1. Interaction-local operations use the interaction guild ID where appropriate.
2. Configured development guild behavior is distinct from production guild behavior.
3. Required intents are explicit and tested; unused intents are not added.
4. Hardcoded guild/user/channel IDs are eliminated or intentionally isolated/configured.
5. Audit logging produces one final record with success/error and non-zero duration.
6. Malformed or incomplete resolved Discord options do not panic or break command execution.
7. Role and mentionable options are represented safely in audit output.

For each behavior, write and run a failing test before implementation.

## Discord API constraints

- Respect Discord interaction acknowledgement timing; defer before slow work.
- Do not require privileged intents unless the Developer Portal configuration and actual event usage justify them.
- Prefer REST fetches or interaction-resolved data over broad cache requirements.
- Keep Discord API calls behind small adapter functions so domain services remain Discord-free.
- Do not log bot tokens, URLs with credentials, or sensitive interaction content.

## Verification

```bash
go test -count=1 ./internal/discord ./internal/logger ./cmd/betrayal-bot
 go test -race -count=1 ./internal/discord ./internal/logger ./cmd/betrayal-bot
go vet ./...
```

Then run the complete global gate. Manually inspect all Discord-facing diffs.

Commit locally only. Do not push or merge.
