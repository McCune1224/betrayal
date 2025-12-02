package logger

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/rs/zerolog"
	"github.com/zekrotja/ken"
)

// CommandAudit represents a command execution record
type CommandAudit struct {
	CorrelationID    string                 `json:"correlation_id"`
	CommandName      string                 `json:"command_name"`
	UserID           string                 `json:"user_id"`
	Username         string                 `json:"username"`
	UserRoles        []string               `json:"user_roles"`
	GuildID          string                 `json:"guild_id"`
	ChannelID        string                 `json:"channel_id"`
	IsAdmin          bool                   `json:"is_admin"`
	CommandArguments map[string]interface{} `json:"command_arguments"`
	Status           string                 `json:"status"` // 'success', 'error', 'cancelled'
	ErrorMessage     *string                `json:"error_message,omitempty"`
	ExecutionTimeMs  int64                  `json:"execution_time_ms"`
	Environment      string                 `json:"environment"`
}

// AuditWriter handles async writing of command audits to the database
type AuditWriter struct {
	pool        *pgxpool.Pool
	channel     chan CommandAudit
	done        chan struct{}
	wg          sync.WaitGroup
	batchSize   int
	flushTimer  *time.Ticker
	environment string
}

// NewAuditWriter creates a new audit writer with async batching
func NewAuditWriter(pool *pgxpool.Pool, environment string) *AuditWriter {
	if pool == nil {
		return nil // Audit writer is optional
	}

	aw := &AuditWriter{
		pool:        pool,
		channel:     make(chan CommandAudit, 100), // Buffered channel to avoid blocking
		done:        make(chan struct{}),
		batchSize:   50,
		flushTimer:  time.NewTicker(3 * time.Second),
		environment: environment,
	}

	// Start background worker goroutine
	aw.wg.Add(1)
	go aw.batchWorker()

	return aw
}

// LogCommand enqueues a command audit for database insertion
func (aw *AuditWriter) LogCommand(audit CommandAudit) {
	if aw == nil || aw.pool == nil {
		return
	}

	audit.Environment = aw.environment

	// Non-blocking send to channel
	select {
	case aw.channel <- audit:
	case <-aw.done:
		return
	default:
		// Channel full, skip to avoid blocking
	}
}

// batchWorker accumulates audit entries and inserts them in batches
func (aw *AuditWriter) batchWorker() {
	defer aw.wg.Done()

	batch := make([]CommandAudit, 0, aw.batchSize)

	for {
		select {
		case audit := <-aw.channel:
			batch = append(batch, audit)
			if len(batch) >= aw.batchSize {
				aw.insertBatch(batch)
				batch = batch[:0]
			}

		case <-aw.flushTimer.C:
			if len(batch) > 0 {
				aw.insertBatch(batch)
				batch = batch[:0]
			}

		case <-aw.done:
			// Drain remaining entries
			close(aw.channel)
			for audit := range aw.channel {
				batch = append(batch, audit)
			}
			if len(batch) > 0 {
				aw.insertBatch(batch)
			}
			aw.flushTimer.Stop()
			return
		}
	}
}

// insertBatch performs a batch insert of audit entries into the database
func (aw *AuditWriter) insertBatch(batch []CommandAudit) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, audit := range batch {
		argumentsJSON, _ := json.Marshal(audit.CommandArguments)

		query := `
			INSERT INTO command_audit (
				correlation_id, command_name, user_id, username, user_roles,
				guild_id, channel_id, is_admin, command_arguments,
				status, error_message, execution_time_ms, environment
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`

		if err := aw.pool.QueryRow(ctx, query,
			audit.CorrelationID,
			audit.CommandName,
			audit.UserID,
			audit.Username,
			audit.UserRoles,
			audit.GuildID,
			audit.ChannelID,
			audit.IsAdmin,
			argumentsJSON,
			audit.Status,
			audit.ErrorMessage,
			audit.ExecutionTimeMs,
			audit.Environment,
		).Scan(); err != nil && err.Error() != "no rows in result set" {
			// Log to stderr on failure (avoid infinite loop)
			zerolog.DefaultContextLogger.Error().Err(err).Msg("Failed to insert command audit")
		}
	}
}

// Close gracefully shuts down the audit writer and flushes pending audits
func (aw *AuditWriter) Close() error {
	if aw == nil || aw.pool == nil {
		return nil
	}

	close(aw.done)
	aw.wg.Wait()
	return nil
}

// ExtractCommandArguments converts Ken options to a map for audit logging
func ExtractCommandArguments(session *discordgo.Session, options []*discordgo.ApplicationCommandInteractionDataOption) map[string]interface{} {
	result := make(map[string]interface{})
	extractOptions(session, result, options, "")
	return result
}

// extractOptions recursively extracts command options and subcommands
func extractOptions(session *discordgo.Session, result map[string]interface{}, options []*discordgo.ApplicationCommandInteractionDataOption, prefix string) {
	for _, opt := range options {
		key := opt.Name
		if prefix != "" {
			key = prefix + "." + opt.Name
		}

		switch opt.Type {
		case discordgo.ApplicationCommandOptionSubCommand:
			result[key] = "subcommand"
			extractOptions(session, result, opt.Options, key)
		case discordgo.ApplicationCommandOptionSubCommandGroup:
			result[key] = "subcommand_group"
			extractOptions(session, result, opt.Options, key)
		case discordgo.ApplicationCommandOptionString:
			result[key] = opt.StringValue()
		case discordgo.ApplicationCommandOptionInteger:
			result[key] = opt.IntValue()
		case discordgo.ApplicationCommandOptionNumber:
			result[key] = opt.FloatValue()
		case discordgo.ApplicationCommandOptionBoolean:
			result[key] = opt.BoolValue()
		case discordgo.ApplicationCommandOptionUser:
			if session != nil {
				result[key] = map[string]interface{}{
					"id":       opt.UserValue(session).ID,
					"username": opt.UserValue(session).Username,
				}
			} else {
				result[key] = "user (unavailable)"
			}
		case discordgo.ApplicationCommandOptionChannel:
			if session != nil {
				result[key] = map[string]interface{}{
					"id":   opt.ChannelValue(session).ID,
					"name": opt.ChannelValue(session).Name,
				}
			} else {
				result[key] = "channel (unavailable)"
			}
		case discordgo.ApplicationCommandOptionRole:
			if session != nil {
				result[key] = map[string]interface{}{
					"id":   opt.RoleValue(session, "").ID,
					"name": opt.RoleValue(session, "").Name,
				}
			} else {
				result[key] = "role (unavailable)"
			}
		case discordgo.ApplicationCommandOptionMentionable:
			result[key] = opt.StringValue()
		default:
			result[key] = "unknown"
		}
	}
}

// CreateAuditFromContext builds a CommandAudit from Ken context
func CreateAuditFromContext(ctx *ken.Ctx, session *discordgo.Session, startTime time.Time) CommandAudit {
	event := ctx.GetEvent()
	execution_time := time.Since(startTime).Milliseconds()

	userRoles := []string{}
	if event.Member != nil {
		userRoles = event.Member.Roles
	}

	isAdmin := false
	if event.Member != nil {
		for _, roleID := range event.Member.Roles {
			for _, adminRole := range discord.AdminRoles {
				if roleID == adminRole {
					isAdmin = true
					break
				}
			}
		}
	}

	cmdData := event.ApplicationCommandData()
	arguments := ExtractCommandArguments(session, cmdData.Options)

	return CommandAudit{
		CorrelationID:    GenerateCorrelationID().String(),
		CommandName:      cmdData.Name,
		UserID:           event.Member.User.ID,
		Username:         event.Member.User.Username,
		UserRoles:        userRoles,
		GuildID:          event.GuildID,
		ChannelID:        event.ChannelID,
		IsAdmin:          isAdmin,
		CommandArguments: arguments,
		Status:           "success",
		ExecutionTimeMs:  execution_time,
	}
}

// Global audit writer instance
var defaultAuditWriter *AuditWriter

// InitAuditWriter initializes the global audit writer
func InitAuditWriter(pool *pgxpool.Pool, env string) {
	defaultAuditWriter = NewAuditWriter(pool, env)
}

// GetAuditWriter returns the global audit writer instance
func GetAuditWriter() *AuditWriter {
	return defaultAuditWriter
}

// CloseAuditWriter closes the global audit writer
func CloseAuditWriter() error {
	if defaultAuditWriter != nil {
		return defaultAuditWriter.Close()
	}
	return nil
}
