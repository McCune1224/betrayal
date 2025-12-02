# Agent Guidelines for Betrayal Bot

## Build & Run Commands

- **Run bot**: `make run` or `go run ./cmd/betrayal-bot/main.go`
- **Run tests**: `go test ./...`
- **Run single test**: `go test -run TestName ./path/to/package`
- **Database migrations**: `make migrate-up`, `make migrate-down`, `make migrate-sync`
- **Mock database**: `make mock-migrate-up`, `make mock-migrate-down`

## Critical Dependencies

**Ken Framework**: https://github.com/zekroTJA/ken - VITAL package for Discord slash command routing and management. All command handlers must implement `ken.Command` interface and be registered via `ken.Ken`.

## Code Style Guidelines

**Language**: Go 1.22+ (see go.mod for exact version)

**Imports**: Organized in three groups separated by blank lines:
1. Standard library (e.g., `context`, `fmt`, `log`)
2. External packages (e.g., `github.com/bwmarrin/discordgo`)
3. Internal packages (e.g., `github.com/mccune1224/betrayal/internal/...`)

**Naming**: PascalCase for types/interfaces, camelCase for functions/variables. Command structs implement `ken.Command` interface.

**Error Handling**: Use `util.ErrorContains(err, msg)` and `util.ErrorNotFound(err)` helpers. Log errors with `log.Printf()` and return descriptive messages to Discord.

**Type Assertions**: Use explicit interface implementations (e.g., `var _ ken.SlashCommand = (*Action)(nil)`)

**Database**: Use `pgx/v5` with connection pools (`*pgxpool.Pool`). Database queries via sqlc in `internal/models/` (SQL-generated Go code).

**Testing**: Use testify suite pattern (`suite.Suite`) with setup in `SetupTest()`. Place tests in `tests/` directory mirroring structure.

**Comments**: Use `//` for single-line, document public functions/types. Include TODOs for incomplete features.

**Package Structure**: Command handlers in `internal/commands/{name}/`, services in `internal/services/`, database layer in `internal/models/` and `internal/db/`

## Git & Build Management

**Binaries**: DO NOT commit binary files to git (e.g., `audit-analysis`, `data-entry`, `betrayal-bot`). Binaries waste repository space and should never be version controlled. Users should build binaries locally with `go build ./cmd/{tool-name}` or use `make` targets. If binaries appear in `git status`, do NOT stage or commit them - delete them locally with `rm {binary-name}`.
