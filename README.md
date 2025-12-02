# Betrayal Discord Bot 🤖

[![Go](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Welcome to the Betrayal Discord Bot - your go-to companion for immersive Betrayal game experiences on Discord! 🎮

## Overview 🌐

The Betrayal Discord Bot, crafted with ❤️ in Go, is designed to assist with the game management for the game Betrayal. This dynamic bot seamlessly integrates with Discord, managing game events, characters, and more, all powered by a robust PostgreSQL database.

## Features 🚀

- **Immersive Game Events:** From inventory creation and management to alliance creations, and near-instant response time, the bot enriches Betrayal game events.
- **Persistent Data Storage:** Utilizes PostgreSQL for storing crucial game data, ensuring a persistent and immersive gaming experience.
- **Timed/Schedule Events:** Manages timed events to trigger at user-provided times to allow for automated and quality of life processes to complete/trigger at the desired time.
- **Discord Magic:** Handles Discord-specific events like slash commands to handle game events and interactions for players, channel creation, and dynamic channel management tailored for the Betrayal game.

## About Betrayal ℹ️
Betrayal is a battle royale game where everyone has unique roles and abilities. Use your abilities and items, allies, and whatever tools at your disposal to be the last one standing. 🏆

## Logging System 📝

The Betrayal bot uses a production-grade structured logging system powered by [zerolog](https://github.com/rs/zerolog).

### Features

- **Structured Logging:** All logs are JSON-formatted for easy parsing and analysis
- **Environment-Based Configuration:**
  - **Local:** Debug level logging with human-readable console output
  - **Staging/Production:** Info level logging to both console and PostgreSQL
- **Correlation IDs:** Every request gets a unique UUID for tracing across the system
- **Async Database Writes:** Non-blocking batch insertion of logs (buffered by 100 logs or 5 seconds)
- **Panic Recovery:** Goroutines wrapped with automatic panic recovery and logging
- **Retention Policy:** Logs retained for 90 days with automatic CSV archival to `./logs_archive/`

### Configuration

Set the `ENVIRONMENT` environment variable to control logging behavior:

```bash
# Development (debug output, no database)
export ENVIRONMENT=local

# Production (info level, database logging)
export ENVIRONMENT=production
```

### Usage Examples

#### Basic Logging

```go
import "github.com/mccune1224/betrayal/internal/logger"

log := logger.Get()
log.Info().Str("user", "alice").Msg("user joined game")
log.Error().Err(err).Msg("failed to process command")
```

#### In Ken Command Handlers

```go
import (
    "github.com/mccune1224/betrayal/internal/logger"
    "github.com/zekrotja/ken"
)

func (cmd *MyCommand) Run(ctx ken.Context) error {
    log := logger.FromKenContext(ctx)
    logger.InjectKenContext(ctx.(*ken.Ctx))
    
    log.Info().Msg("command executed")
    return nil
}
```

#### Safe Goroutines

```go
log := logger.Get()

// Launches goroutine with panic recovery and error logging
logger.SafeGo(*log, "background_task", func() error {
    // Your async work here
    return nil
})

// For cleanup operations that don't return errors
logger.SafeGoVoid(*log, "cleanup_task", func() {
    // Your cleanup here
})
```

### Database Schema

Logs are stored in the `logs` table with the following fields:

- `id` - Auto-incrementing primary key
- `timestamp` - When the log entry was created
- `level` - Log level (debug, info, warn, error, fatal, panic)
- `message` - Main log message
- `correlation_id` - UUID for request tracing
- `user_id` - Discord user ID (if applicable)
- `command_name` - Discord command name (if applicable)
- `error_details` - JSON object with error information
- `request_data` - JSON object with additional context
- `environment` - Deployment environment (local, staging, production)
- `created_at` - When the log was inserted

### Querying Logs

```sql
-- Get all errors from last hour
SELECT * FROM logs
WHERE level = 'error'
AND created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

-- Trace a specific user's actions
SELECT * FROM logs
WHERE user_id = '206268866714796032'
ORDER BY created_at DESC
LIMIT 100;

-- Find all logs for a correlation ID
SELECT * FROM logs
WHERE correlation_id = '550e8400-e29b-41d4-a716-446655440000'
ORDER BY created_at DESC;
```

### Architecture

The logger system consists of:

- **logger/logger.go** - Core initialization and logger instance management
- **logger/database.go** - Async batch writer for PostgreSQL (non-blocking)
- **logger/middleware.go** - Correlation ID generation and Ken context integration
- **logger/recovery.go** - Panic recovery wrappers for handlers and commands
- **logger/goroutines.go** - SafeGo wrappers for protected goroutine execution
- **logger/retention.go** - Automatic 90-day retention and CSV archival

## Command Audit Trail 📋

The bot maintains a comprehensive audit trail of all player and admin commands for tracking and archival purposes.

### Features

- **Complete Command Tracking:** Every slash command execution is logged with full arguments
- **User Attribution:** Tracks which user executed the command and their roles
- **Admin Detection:** Automatically identifies admin vs. player commands
- **Argument Preservation:** All command arguments captured in JSON format for easy analysis
- **Error Logging:** Failed commands logged with error messages for debugging
- **Execution Timing:** Records how long each command took to execute
- **Async Batch Writes:** Non-blocking audit logging (buffered by 50 entries or 3 seconds)

### Command Audit Table Schema

The `command_audit` table stores:

- `id` - Auto-incrementing primary key
- `correlation_id` - Links to log entries for full request tracing
- `timestamp` - When the command was executed
- `command_name` - Name of the slash command (e.g., "inv", "vote")
- `user_id` - Discord user ID who executed the command
- `username` - Discord username for easy identification
- `user_roles` - Array of roles the user had at execution time
- `guild_id` - Guild where command was executed
- `channel_id` - Channel where command was executed
- `is_admin` - Boolean flag for admin commands
- `command_arguments` - JSONB with all command arguments/options
- `status` - Execution status: 'success', 'error', 'cancelled'
- `error_message` - Error message if status is 'error'
- `execution_time_ms` - Milliseconds taken to execute command
- `environment` - Deployment environment (local, staging, production)

### Querying Audit Data

```sql
-- Find all commands executed by a specific user
SELECT command_name, command_arguments, status, timestamp
FROM command_audit
WHERE user_id = '206268866714796032'
ORDER BY timestamp DESC
LIMIT 50;

-- Track admin commands from the last 24 hours
SELECT username, command_name, command_arguments, status, timestamp
FROM command_audit
WHERE is_admin = true
AND timestamp > NOW() - INTERVAL '24 hours'
ORDER BY timestamp DESC;

-- Find failed commands for debugging
SELECT command_name, username, error_message, timestamp
FROM command_audit
WHERE status = 'error'
AND timestamp > NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC;

-- Audit who used a specific command
SELECT DISTINCT username, timestamp, command_arguments
FROM command_audit
WHERE command_name = 'inv'
ORDER BY timestamp DESC;

-- Performance analysis - slowest commands
SELECT command_name, AVG(execution_time_ms) as avg_time, MAX(execution_time_ms) as max_time, COUNT(*) as count
FROM command_audit
WHERE timestamp > NOW() - INTERVAL '7 days'
GROUP BY command_name
ORDER BY avg_time DESC;

-- Trace a user's actions via correlation ID
SELECT l.timestamp, l.level, l.message, ca.command_name, ca.status
FROM logs l
LEFT JOIN command_audit ca ON l.correlation_id = ca.correlation_id
WHERE ca.user_id = '206268866714796032'
ORDER BY l.timestamp DESC;
```

### Example Command Arguments in Audit

Different command types store their arguments:

**Inventory Command:**
```json
{
  "action": "view",
  "player": { "id": "123456789", "username": "alice" }
}
```

**Vote Command:**
```json
{
  "action": "vote",
  "target": { "id": "987654321", "username": "bob" }
}
```

**Buy Command:**
```json
{
  "item": "health_potion"
}
```

### Access Control

Command audit is automatically integrated:
- **Automatic Logging:** All commands are logged by default
- **No Setup Required:** Works immediately after migration
- **Async Processing:** Non-blocking, doesn't slow down command execution
- **Error Resilience:** Failed audit writes don't break command execution

### Use Cases

1. **Compliance:** Maintain audit trail for game rule violations
2. **Debugging:** Understand what command arguments caused errors
3. **Analytics:** Identify most-used features and performance bottlenecks
4. **Fraud Detection:** Track unusual command patterns or admin abuse
5. **Game Balance:** Analyze which items/abilities are most purchased/used

## License 📄
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
