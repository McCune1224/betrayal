# Betrayal Bot - Channel Configuration Quick Reference

## Channel Types at a Glance

### 1. Admin Channels (Multiple)
- **Purpose**: Whitelist channels for `/inv` command usage outside confessionals
- **Command**: `/channel admin [add|list|delete] [channel]`
- **DB Table**: `admin_channel` (channel_id VARCHAR UNIQUE)
- **File**: `internal/commands/channels/admin.go` (lines 14-114)
- **Note**: Can have multiple admin channels

### 2. Vote Channel (Single)
- **Purpose**: Funnel channel where players submit votes
- **Command**: `/channel vote [update|view] [channel]`
- **DB Table**: `vote_channel` (channel_id VARCHAR UNIQUE)
- **File**: `internal/commands/channels/vote.go` (lines 14-81)
- **Note**: Used in `/cycle` broadcasts; UPSERT pattern (replaces on update)

### 3. Action Channel (Single)
- **Purpose**: Funnel channel where players submit actions
- **Command**: `/channel action [update|view] [channel]`
- **DB Table**: `action_channel` (channel_id VARCHAR UNIQUE)
- **File**: `internal/commands/channels/action.go` (lines 14-88)
- **Note**: Used in `/cycle` broadcasts; WIPE+UPSERT pattern

### 4. Lifeboard (Single)
- **Purpose**: Player status board (alive/dead display with sorting)
- **Command**: `/channel lifeboard set [channel]`
- **DB Table**: `player_lifeboard` (channel_id, message_id VARCHAR UNIQUE)
- **File**: `internal/commands/channels/lifeboard.go` (lines 17-135)
- **Behavior**:
  - Pins message to channel
  - Sorts: alive players first (A-Z), then dead players (A-Z)
  - Includes EST timestamp footer
  - Updates existing lifeboard when set again

### 5. Confessionals (Multiple, 1 per player)
- **Purpose**: Private per-player channels for admin communication
- **Command**: `/channel confessionals` (view only)
- **DB Table**: `player_confessional` (player_id, channel_id, pin_message_id)
- **File**: `internal/commands/channels/channels.go` (lines 70-91)
- **Note**: Created during game setup, not via `/channel` command

---

## Configuration Setup Order (Recommended)

1. **Create Confessionals** (per player - outside `/channel` command)
2. **Set Vote Channel**: `/channel vote update #voting-funnel`
3. **Set Action Channel**: `/channel action update #action-funnel`
4. **Add Admin Channel(s)**: `/channel admin add #admin-operations`
5. **Set Lifeboard**: `/channel lifeboard set #status-board`
6. **Verify Setup**: 
   - `/channel confessionals` (view all)
   - `/channel vote view`
   - `/channel action view`
   - `/channel admin list`

---

## Admin Role Definition

**Location**: `internal/discord/role.go` (lines 9-13)

**Roles with Admin Access**:
- "Host"
- "Co-Host"
- "Bot Developer"

---

## Channel Broadcast Flow (Cycle)

When `/cycle next` or `/cycle set` is executed:

1. Gets all confessional channel IDs
2. Gets vote channel ID
3. Gets action channel ID
4. Gets all alliance channel IDs (from category)
5. Sends cycle message to ALL of the above

**Location**: `internal/commands/cycle/cycle.go` (lines 200-235)

---

## Database Operations Quick Reference

### Admin Channels
```go
q.ListAdminChannel(ctx)              // []string
q.CreateAdminChannel(ctx, channelID) // string
q.DeleteAdminChannel(ctx, channelID) // error
```

### Vote Channel
```go
q.GetVoteChannel(ctx)            // string
q.UpsertVoteChannel(ctx, ch)     // error
q.WipeVoteChannel(ctx)           // error
```

### Action Channel
```go
q.GetActionChannel(ctx)          // string
q.UpsertActionChannel(ctx, ch)   // error
q.WipeActionChannel(ctx)         // error
```

### Lifeboard
```go
q.CreatePlayerLifeboard(ctx, params)  // PlayerLifeboard
q.GetPlayerLifeboard(ctx)             // PlayerLifeboard
q.DeletePlayerLifeboard(ctx)          // error
```

### Confessionals
```go
q.ListPlayerConfessional(ctx)                           // []PlayerConfessional
q.CreatePlayerConfessional(ctx, playerID, chID, msgID) // PlayerConfessional
q.GetPlayerConfessional(ctx, playerID)                 // PlayerConfessional
q.GetPlayerConfessionalByChannelID(ctx, chID)          // PlayerConfessional
q.DeletePlayerConfessional(ctx, playerID)              // error
```

---

## Command Auth Pattern

All channel commands check:
```go
if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
    return discord.NotAdminError(ctx)
}
```

**Files with checks**:
- `admin.go` (lines 58, 74, 90)
- `vote.go` (lines 49, 65)
- `action.go` (lines 50, 74)

---

## Common Issues & Solutions

### Issue: Vote/Action channel "not found" errors
**Solution**: Use `/channel vote view` or `/channel action view` to verify they're set

### Issue: Confessional not receiving cycle messages
**Solution**: Run `/channel confessionals` to list all; check if missing any

### Issue: Lifeboard message doesn't update
**Solution**: Re-run `/channel lifeboard set #channel` to force update

### Issue: Admin commands failing
**Solution**: Verify user has "Host", "Co-Host", or "Bot Developer" role

---

## File Location Reference

### Commands
- Main: `internal/commands/channels/channels.go`
- Admin: `internal/commands/channels/admin.go`
- Vote: `internal/commands/channels/vote.go`
- Action: `internal/commands/channels/action.go`
- Lifeboard: `internal/commands/channels/lifeboard.go`

### Database Schema
- Admin: `internal/db/migration/000018_admin_channels.up.sql`
- Vote: `internal/db/migration/000019_vote_channel.up.sql`
- Action: `internal/db/migration/000020_action_channel.up.sql`
- Lifeboard: `internal/db/migration/000021_player_lifeboard.up.sql`
- Confessional: `internal/db/migration/000016_player_confessional.sql.up.sql`

### Generated Models
- Admin: `internal/models/admin_channel.sql.go`
- Vote: `internal/models/vote_channel.sql.go`
- Action: `internal/models/action_channel.sql.go`
- Lifeboard: `internal/models/player_lifeboard.sql.go`
- Confessional: `internal/models/player_confessional.sql.go`

### Help Documentation
- Admin help: `internal/commands/help/admin.go` (lines 10-216)
- Help messages: `internal/commands/help/admin_messages.go` (lines 143-172)

---

## Missing Features (Recommended Additions)

1. `/admin health` - Verify all channels exist in Discord
2. `/admin status` - Show configuration state
3. Configuration validation before game start
4. Orphaned channel detection
5. Channel recovery procedures
6. Error messages when channels are deleted mid-game
