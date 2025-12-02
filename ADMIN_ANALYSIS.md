# Betrayal Bot Admin Setup & Configuration Analysis

## 1. CURRENT ADMIN SETUP HELP DOCUMENTATION

### Location & Files
- **Main Admin Help Handler**: `/home/mckusa/Code/betrayal/internal/commands/help/admin.go` (Lines 10-216)
- **Admin Help Messages/Embeds**: `/home/mckusa/Code/betrayal/internal/commands/help/admin_messages.go` (Lines 143-172)

### What IS Currently Documented

**Admin Overview** (`adminOverview`, admin.go:10-151):
- Shows 8 main admin command categories with brief descriptions:
  1. Inventory (`/inv`)
  2. Alliance (`/alliance admin`)
  3. Channels (`/channel`)
  4. Cycle (`/cycle`)
  5. Roll (`/roll`)
  6. Buy (`/buy`)
  7. Kill/Revive (`/kill`, `/revive`)
  8. Setup (`/setup`)

**Channels Help** (`adminChannelsEmbed`, admin_messages.go:143-172):
- Admin Channel: add/delete/list commands
- Vote Channels: add/delete/list commands
- Action Channels: add/delete/list commands
- Lifeboard: set/delete/list commands
- Confessionals: view all current player confessional channels

**Inventory Help** (admin_messages.go:7-40):
- Basic modifications format: `/inv [category] [add/remove/set] [player]`
- Categories listed: ability, item, coin, status, perk, alignment, role, immunity, luck, death, notes
- Timed effects support (status/perks with duration: "1h30m", "45m", "2h")
- Inventory creation/deletion
- Whitelist management for inventory commands

**Cycle Help** (admin_messages.go:174-199):
- Current phase viewing
- Advancing to next phase
- Manual phase setting: `/cycle set [phase] [number]`
- Broadcasting to confessionals, alliances, and funnel channels

**Other Documented Commands** (admin_messages.go):
- Setup (role list generation)
- Buy (purchase items on behalf)
- Kill/Revive (player death management, lifeboard)
- Roll (manual rolls, event rolls, wheel)
- Alliance (creation/invite approval, wipe)

### What IS MISSING or UNDERSPECIFIED

1. **Channel Configuration Setup Flow** - No guidance on which channels to set in what order for a new game
2. **Initial Game Setup Checklist** - No document listing all configuration steps needed before starting a game
3. **Channel Dependencies** - No explanation of how channels relate to each other (e.g., vote/action are funnel channels)
4. **Admin Channel Purpose** - Listed but not explained (it's for inventory command whitelisting)
5. **Lifeboard vs Kill Location** - Two different concepts but not clearly distinguished
6. **Error Handling & Recovery** - No documentation on what to do if channels are missing/deleted
7. **Multiple Admin Channels** - Can add multiple admin channels but help doesn't clarify purpose
8. **Vote/Action Channel Limits** - Single or multiple? Only vote is documented as "single"
9. **Database Schema Context** - No mention of backend tables/structure for admin reference
10. **Health/Status Checks** - No commands for checking if configuration is complete/valid

---

## 2. CHANNEL TYPES & CONFIGURATION

### Channel Management Command Structure
**Location**: `/home/mckusa/Code/betrayal/internal/commands/channels/`

#### Sub-commands Available:
```
/channel [subgroup] [command] [options]
├── /channel admin [add/list/delete] [channel]
├── /channel vote [update/view] [channel]
├── /channel action [update/view] [channel]
├── /channel lifeboard [set] [channel]
└── /channel confessionals (view current)
```

### Channel Types & Details

#### 1. **ADMIN CHANNELS** (Multiple allowed)
- **File**: `admin.go` (Lines 14-114)
- **Commands**:
  - `add [channel]` - Create admin_channel record (line 53-67)
  - `list` - View all admin channels (line 85-114)
  - `delete [channel]` - Remove admin_channel record (line 69-83)
- **Purpose**: Whitelist for inventory commands (allows `/inv` usage outside confessionals)
- **Auth**: Requires `IsAdminRole(ctx, discord.AdminRoles...)` checks (lines 58, 74, 90)
- **Database**: `admin_channel` table, stores `channel_id` (VARCHAR)

#### 2. **VOTE CHANNEL** (Single)
- **File**: `vote.go` (Lines 14-81)
- **Commands**:
  - `update [channel]` - Upsert vote_channel record (line 44-58)
  - `view` - Show current vote channel (line 60-81)
- **Purpose**: Funnel channel where players submit votes (used in `/cycle` broadcasts)
- **Auth**: Requires admin role
- **Database**: `vote_channel` table, single row, stores `channel_id` (VARCHAR)
- **Note**: Uses UPSERT pattern (only one vote channel at a time)

#### 3. **ACTION CHANNEL** (Single)
- **File**: `action.go` (Lines 14-88)
- **Commands**:
  - `update [channel]` - Wipe old, then upsert new action_channel (line 69-88)
  - `view` - Show current action channel (line 45-67)
- **Purpose**: Funnel channel where players submit actions (used in `/cycle` broadcasts)
- **Auth**: Requires admin role
- **Database**: `action_channel` table, single row, stores `channel_id` (VARCHAR)
- **Note**: Uses WIPE + UPSERT pattern on update (line 80)

#### 4. **LIFEBOARD** (Single)
- **File**: `lifeboard.go` (Lines 17-135)
- **Commands**:
  - `set [channel]` - Create/update lifeboard with pinned message (line 41-86)
- **Purpose**: Display player status board (alive/dead status, sorted)
- **Database**: `player_lifeboard` table, stores `channel_id` and `message_id` (both VARCHAR)
- **Behavior** (line 52-55):
  - Deletes old lifeboard message and database record
  - Builds new message with player statuses
  - Sends message to channel, pins it, stores metadata
  - Message sorted: alive players first (alphabetically), then dead players (alphabetically)
  - Includes footer with EST timestamp

#### 5. **CONFESSIONALS** (Multiple, one per player)
- **File**: `channels.go` (Lines 70-91)
- **Command**: `confessionals` (view only)
- **Purpose**: Private channels for individual players to communicate with admins/view inventory
- **Database**: `player_confessional` table
  - Fields: `player_id` (BIGINT), `channel_id` (BIGINT), `pin_message_id` (BIGINT)
  - Primary Key: (player_id, channel_id)
- **Created By**: Implied to be created during game setup (not in `/channel` commands)
- **Used By**: Cycle broadcasts, inventory commands, player communications

---

## 3. DATABASE TABLES & QUERIES FOR CHANNEL CONFIGURATION

### Migration Files & Schema

| Migration # | Table Name | Purpose | Fields |
|------------|-----------|---------|--------|
| 000018 | `admin_channel` | Whitelisted admin channels | `channel_id` (VARCHAR, UNIQUE) |
| 000019 | `vote_channel` | Single voting funnel | `channel_id` (VARCHAR, UNIQUE) |
| 000020 | `action_channel` | Single action funnel | `channel_id` (VARCHAR, UNIQUE) |
| 000021 | `player_lifeboard` | Player status board | `channel_id` (VARCHAR, UNIQUE), `message_id` (VARCHAR, UNIQUE) |
| 000016 | `player_confessional` | Player private channels | `player_id` (BIGINT ref), `channel_id` (BIGINT), `pin_message_id` (BIGINT), PRIMARY KEY(player_id, channel_id) |

### Query Files & Generated Go Code

**Location**: `/home/mckusa/Code/betrayal/internal/db/query/` (SQL definitions)
**Generated**: `/home/mckusa/Code/betrayal/internal/models/` (*.sql.go files with Go functions)

#### Admin Channel Queries
- **File**: `admin_channel.sql` (lines 1-17)
- **Generated**: `admin_channel.sql.go` (lines 1-60)
- **Operations**:
  - `CreateAdminChannel(ctx, channelID)` → Returns created channel_id
  - `ListAdminChannel(ctx)` → Returns []string of all channel IDs
  - `DeleteAdminChannel(ctx, channelID)` → Deletes by channel_id

#### Vote Channel Queries
- **File**: `vote_channel.sql` (implicit)
- **Generated**: `vote_channel.sql.go` (lines 1-43)
- **Operations**:
  - `GetVoteChannel(ctx)` → Returns single channel_id
  - `UpsertVoteChannel(ctx, channelID)` → Insert/update single record
  - `WipeVoteChannel(ctx)` → Delete all (prepare for new)

#### Action Channel Queries
- **File**: `action_channel.sql` (implicit)
- **Generated**: `action_channel.sql.go` (lines 1-43)
- **Operations**:
  - `GetActionChannel(ctx)` → Returns single channel_id
  - `UpsertActionChannel(ctx, channelID)` → Insert/update single record
  - `WipeActionChannel(ctx)` → Delete all (prepare for new)

#### Lifeboard Queries
- **Generated**: `player_lifeboard.sql.go` (lines 1-89)
- **Operations**:
  - `CreatePlayerLifeboard(ctx, params)` → Create/upsert lifeboard record
  - `GetPlayerLifeboard(ctx)` → Get current lifeboard (channel + message IDs)
  - `DeletePlayerLifeboard(ctx)` → Delete lifeboard record

#### Confessional Queries
- **Generated**: `player_confessional.sql.go` (lines 1-89)
- **Operations**:
  - `CreatePlayerConfessional(ctx, playerID, channelID, pinMessageID)` → Create confession
  - `GetPlayerConfessional(ctx, playerID)` → Get confessional for player
  - `GetPlayerConfessionalByChannelID(ctx, channelID)` → Get player for channel
  - `ListPlayerConfessional(ctx)` → Get all (used in cycle broadcasts)
  - `DeletePlayerConfessional(ctx, playerID)` → Remove confessional

#### Cycle Queries (used in broadcasts)
- **Generated**: `cycle.sql.go` (lines 1-45)
- **Operations**:
  - `GetCycle(ctx)` → Get current phase (Day/Elimination, number)
  - `UpdateCycle(ctx, params)` → Set new phase

---

## 4. EXISTING HEALTH/STATUS CHECKING MECHANISMS

### Current Status Monitoring

#### A. **Command Audit Trail** (Operational Health)
- **Location**: `/home/mckusa/Code/betrayal/internal/logger/audit_commands.go`
- **Type**: Async batch logging to `command_audit` table
- **Records Captured** (CommandAudit struct, audit_commands.go:17-31):
  - `correlation_id` - UUID for tracing
  - `command_name` - Slash command name
  - `user_id`, `username`, `user_roles` - Who ran it
  - `guild_id`, `channel_id` - Where it ran
  - `is_admin` - Admin flag
  - `command_arguments` - Arguments passed
  - `status` - 'success', 'error', 'cancelled'
  - `error_message` - Error details
  - `execution_time_ms` - Performance tracking
  - `environment` - Deployment environment

**Database**: `command_audit` table (migration 000025)
- Retention: 365 days (cleaned daily via `CleanupAuditTrail()`)

#### B. **Panic Recovery & Error Logging**
- **Location**: `/home/mckusa/Code/betrayal/internal/logger/recovery.go`
- **Functions**:
  - `RecoverWithLog()` - Generic panic recovery (lines 11-24)
  - `RecoverKenCommand()` - Command handler panic recovery (lines 27-38)
  - `RecoverKenComponent()` - Component handler panic recovery (lines 49-71)
  - `WrapKenHandler()`, `WrapKenComponent()` - Wrapper functions
- **Behavior**: Logs panic with stack trace, attempts Discord error response, may re-panic in production

#### C. **Correlation Tracking**
- **Location**: `/home/mckusa/Code/betrayal/internal/logger/middleware.go`
- **Functions**:
  - `GenerateCorrelationID()` - Creates UUID for request tracing (line 61)
  - `InjectKenContext()` - Adds correlation ID to Ken context (line 67)
  - `FromKenContext()` - Extracts context for logging (line 21)
- **Purpose**: Track requests through logs and database for debugging

#### D. **Retention & Archival**
- **Location**: `/home/mckusa/Code/betrayal/internal/logger/retention.go`
- **Features**:
  - `StartRetentionWorker()` - Runs daily at midnight (line 23)
  - `archiveAndCleanLogs()` - Exports old logs to CSV, deletes from DB (line 55)
  - `CleanupAuditTrail()` - Deletes audit records >365 days old (line 197)
  - Archive format: CSV with timestamp, level, message, correlation_id, etc.
  - Retention: 90 days for logs, 365 days for audit (configurable)

#### E. **Database Health Check** (via connection pool)
- **Location**: `/home/mckusa/Code/betrayal/internal/logger/database.go`
- **Available**: pgx/v5 connection pool with built-in health checking
- **Not explicitly exposed**: No `/health` or admin command to check DB status

### WHAT IS MISSING (No Health/Status Checks)

1. **No `/health` or `/status` admin command** - Cannot verify channel configuration
2. **No channel validation** - Cannot check if configured channels still exist in Discord
3. **No configuration completeness check** - Cannot verify all required channels are set
4. **No database connectivity check** - No explicit health endpoint
5. **No cycle/game state validation** - Cannot check if game state is consistent
6. **No orphaned channel detection** - Cannot identify missing confessionals
7. **No message delivery verification** - Cannot check if cycle messages are sending
8. **No recovery procedures** - No documented steps if channels become invalid

---

## 5. COMMAND STRUCTURE & PATTERNS

### Ken Framework Integration
- **Framework**: https://github.com/zekroTJA/ken - Discord slash command routing
- **Pattern**: All commands implement `ken.Command` interface

### Admin Command Pattern

**Location**: `/home/mckusa/Code/betrayal/internal/commands/channels/channels.go`

#### Structure Template
```go
type Channel struct {
    dbPool *pgxpool.Pool
}

// Implement ken.SlashCommand interface
var _ ken.SlashCommand = (*Channel)(nil)

// Required methods:
func (c *Channel) Name() string { return "channel" }
func (c *Channel) Description() string { return "..." }
func (c *Channel) Version() string { return "1.0.0" }
func (c *Channel) Options() []*discordgo.ApplicationCommandOption { /* defines subcommands */ }
func (c *Channel) Run(ctx ken.Context) error { /* routes to subcommands */ }
func (c *Channel) Initialize(pool *pgxpool.Pool) { c.dbPool = pool }
```

#### Subcommand Pattern
```go
// Define subcommand group
func (c *Channel) adminCommandGroupBuilder() ken.SubCommandGroup {
    return ken.SubCommandGroup{
        Name: "admin",
        SubHandler: []ken.CommandHandler{
            ken.SubCommandHandler{Name: "add", Run: c.addAdminChannel},
            ken.SubCommandHandler{Name: "list", Run: c.listAdminChannel},
            ken.SubCommandHandler{Name: "delete", Run: c.deleteAdminChannel},
        },
    }
}

// Define arguments
func (c *Channel) adminCommandArgBuilder() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Type: discordgo.ApplicationCommandOptionSubCommandGroup,
        Name: "admin",
        Options: []*discordgo.ApplicationCommandOption{ /* subcommand defs */ },
    }
}

// Implement subcommand handler
func (c *Channel) addAdminChannel(ctx ken.SubCommandContext) error {
    if err := ctx.Defer(); err != nil { /* defer response */ }
    if !discord.IsAdminRole(ctx, discord.AdminRoles...) { /* check permission */ }
    // ... execute command
}
```

### Admin Role Checking Pattern
- **Location**: `/home/mckusa/Code/betrayal/internal/discord/role.go` (lines 8-46)
- **Defined Roles**: "Host", "Co-Host", "Bot Developer" (lines 9-13)
- **Check Function**: `IsAdminRole(ctx ken.Context, adminRoles ...string) bool` (line 16)
- **Triple-loop Pattern** (lines 21-30):
  1. Check user's roles
  2. Check guild's role definitions
  3. Check against admin roles list

### Error & Success Response Pattern
- **Error Messages**: `discord.AlexError(ctx, msg)`, `discord.ErrorMessage(ctx, title, description)`
- **Success Messages**: `discord.SuccessfulMessage(ctx, title, description)`
- **Logging**: `logger.Get().Error().Err(err).Msg("operation failed")`
- **Context Deferral**: `ctx.Defer()` - allows time for complex operations

### Database Query Pattern
```go
q := models.New(c.dbPool)          // Create query client
dbCtx := context.Background()        // Create context
result, err := q.GetVoteChannel(dbCtx) // Execute query
if err != nil {
    return discord.AlexError(ctx, "error message")
}
```

---

## 6. ADMIN COMMAND INVENTORY

### Currently Implemented Admin Commands
| Command | Subgroup | Subcommand | Purpose |
|---------|----------|-----------|---------|
| `/channel` | `admin` | `add` | Add admin-whitelisted channel |
| `/channel` | `admin` | `list` | List admin channels |
| `/channel` | `admin` | `delete` | Remove admin channel |
| `/channel` | `vote` | `update` | Set vote funnel channel |
| `/channel` | `vote` | `view` | View vote channel |
| `/channel` | `action` | `update` | Set action funnel channel |
| `/channel` | `action` | `view` | View action channel |
| `/channel` | `lifeboard` | `set` | Set lifeboard channel |
| `/channel` | `confessionals` | - | View all confessionals |
| `/cycle` | - | `current` | View current phase |
| `/cycle` | - | `next` | Advance to next phase |
| `/cycle` | - | `set` | Manual phase setting |
| `/setup` | - | - | Generate role list |
| `/inv` | various | various | Inventory management |
| `/kill` | - | - | Mark player dead |
| `/revive` | - | - | Revive player |
| `/buy` | - | - | Purchase item for player |
| `/roll` | various | various | Roll events/items |
| `/alliance` | `admin` | various | Approve alliances |

---

## 7. RECOMMENDATIONS FOR MISSING DOCUMENTATION

### Priority 1: Initial Setup Guide
Create documentation for:
1. Required channels to create (confessionals for each player first)
2. Order of configuration steps
3. Example: "First create confessionals, then set vote/action channels, then set lifeboard"

### Priority 2: Health Check Command
Implement `/admin health` or `/admin status` to:
- Verify all configured channels still exist in Discord
- Check database connectivity
- List configured channels with validation status
- Report missing or orphaned confessionals

### Priority 3: Configuration Validation
Add to admin help:
- What happens if a channel is deleted
- How to recover from missing channels
- Which commands require which channels to be set

### Priority 4: Channel Relationship Documentation
Explain:
- Vote/Action are "funnel" channels used in `/cycle` broadcasts
- Admin channels are separate from confessionals
- Lifeboard is independent message, not tied to other channels
- Confessionals are player-specific, created separately

---

## FILE REFERENCE SUMMARY

| Component | Primary File | Line Range |
|-----------|-------------|-----------|
| Admin Help UI | `commands/help/admin.go` | 10-216 |
| Admin Help Text | `commands/help/admin_messages.go` | 1-200 |
| Channel Commands | `commands/channels/channels.go` | 1-92 |
| Admin Channels | `commands/channels/admin.go` | 1-114 |
| Vote Channel | `commands/channels/vote.go` | 1-81 |
| Action Channel | `commands/channels/action.go` | 1-88 |
| Lifeboard | `commands/channels/lifeboard.go` | 1-135 |
| Cycle Commands | `commands/cycle/cycle.go` | 1-287 |
| Setup Command | `commands/setup/setup.go` | 1-256 |
| DB Schema: Admin | `db/migration/000018_admin_channels.up.sql` | 1-14 |
| DB Schema: Vote | `db/migration/000019_vote_channel.up.sql` | 1-7 |
| DB Schema: Action | `db/migration/000020_action_channel.up.sql` | 1-8 |
| DB Schema: Lifeboard | `db/migration/000021_player_lifeboard.up.sql` | 1-6 |
| DB Schema: Confessional | `db/migration/000016_player_confessional.sql.up.sql` | 1-8 |
| Queries: Admin | `db/query/admin_channel.sql` | 1-17 |
| Queries: Vote | (implicit) | - |
| Queries: Confessional | `db/query/player_confessional.sql` | 1-27 |
| Models: Admin Queries | `models/admin_channel.sql.go` | 1-60 |
| Models: Vote Queries | `models/vote_channel.sql.go` | 1-43 |
| Models: Action Queries | `models/action_channel.sql.go` | 1-43 |
| Models: Lifeboard Queries | `models/player_lifeboard.sql.go` | 1-89 |
| Models: Confessional Queries | `models/player_confessional.sql.go` | 1-89 |
| Logger: Audit | `logger/audit_commands.go` | 1-250+ |
| Logger: Recovery | `logger/recovery.go` | 1-80 |
| Logger: Middleware | `logger/middleware.go` | 1-87 |
| Logger: Retention | `logger/retention.go` | 1-220 |
| Discord Utilities: Role | `discord/role.go` | 1-47 |
