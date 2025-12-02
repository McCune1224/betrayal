package logger

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/zekrotja/ken"
)

// String constants for context keys (using Ken's Set/Get interface on *ken.Ctx only)
const (
	// CorrelationIDKey stores the correlation ID in Ken context
	CorrelationIDKey = "correlation_id"
	// UserIDKey stores the Discord user ID in Ken context
	UserIDKey = "user_id"
	// CommandNameKey stores the command name in Ken context
	CommandNameKey = "command_name"
)

// FromKenContext extracts a logger from Ken context (works with both Ctx and SubCommandContext)
// It attempts to pull stored values and enhance with event data
func FromKenContext(ctx ken.Context) zerolog.Logger {
	logger := defaultLogger

	// Try to cast to *ken.Ctx to access Get method
	if ctxConcrete, ok := ctx.(*ken.Ctx); ok {
		// Inject correlation ID if present
		if corrID, ok := ctxConcrete.Get(CorrelationIDKey).(uuid.UUID); ok {
			logger = logger.With().Str("correlation_id", corrID.String()).Logger()
		}

		// Inject user ID if present
		if userID, ok := ctxConcrete.Get(UserIDKey).(string); ok {
			logger = logger.With().Str("user_id", userID).Logger()
		}

		// Inject command name if present
		if cmdName, ok := ctxConcrete.Get(CommandNameKey).(string); ok {
			logger = logger.With().Str("command", cmdName).Logger()
		}
	}

	// Add command info from event
	if ctx != nil {
		if event := ctx.GetEvent(); event != nil && event.ApplicationCommandData().Name != "" {
			logger = logger.With().Str("command", event.ApplicationCommandData().Name).Logger()
		}

		// Add user info from event
		if event := ctx.GetEvent(); event != nil && event.Member != nil && event.Member.User != nil {
			logger = logger.With().
				Str("user", event.Member.User.Username).
				Str("user_id", event.Member.User.ID).
				Logger()
		}
	}

	return logger
}

// GenerateCorrelationID creates a new correlation ID
func GenerateCorrelationID() uuid.UUID {
	return uuid.New()
}

// InjectKenContext injects correlation ID and command info into Ken context
// Only works with *ken.Ctx (not interfaces), so it's optional
func InjectKenContext(kenCtx *ken.Ctx) {
	if kenCtx == nil {
		return
	}

	// Generate correlation ID if not present
	if kenCtx.Get(CorrelationIDKey) == nil {
		kenCtx.Set(CorrelationIDKey, GenerateCorrelationID())
	}

	// Inject command name
	if kenCtx.GetEvent() != nil && kenCtx.GetEvent().ApplicationCommandData().Name != "" {
		kenCtx.Set(CommandNameKey, kenCtx.GetEvent().ApplicationCommandData().Name)
	}

	// Inject user ID
	if kenCtx.GetEvent() != nil && kenCtx.GetEvent().Member != nil && kenCtx.GetEvent().Member.User != nil {
		kenCtx.Set(UserIDKey, kenCtx.GetEvent().Member.User.ID)
	}
}
