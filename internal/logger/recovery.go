package logger

import (
	"runtime/debug"

	"github.com/rs/zerolog"
	"github.com/zekrotja/ken"
)

// RecoverWithLog is a defer wrapper that catches panics and logs them
func RecoverWithLog(logger zerolog.Logger) {
	if r := recover(); r != nil {
		logger.Error().
			Any("panic", r).
			Bytes("stack", debug.Stack()).
			Msg("Recovered from panic")

		// Re-panic in production for alerting systems to catch
		// In development, just log and continue
		if logger.GetLevel() <= zerolog.InfoLevel {
			panic(r)
		}
	}
}

// RecoverKenCommand catches panics in Ken command handlers and sends error to Discord
func RecoverKenCommand(ctx *ken.Ctx) {
	if r := recover(); r != nil {
		logger := FromKenContext(ctx)
		logger.Error().
			Any("panic", r).
			Bytes("stack", debug.Stack()).
			Msg("Command handler panicked")

		// Try to send error message to Discord
		_ = ctx.RespondError("A fatal error occurred in the command handler", "Error")
	}
}

// WrapKenHandler wraps a Ken command handler with panic recovery
func WrapKenHandler(handler func(*ken.Ctx) error) func(*ken.Ctx) error {
	return func(ctx *ken.Ctx) error {
		defer RecoverKenCommand(ctx)
		return handler(ctx)
	}
}

// RecoverKenComponent catches panics in Ken component handlers and sends error to Discord
func RecoverKenComponent(ctx ken.ComponentContext) {
	if r := recover(); r != nil {
		// For components, we need to extract logger info manually since ComponentContext
		// doesn't inherit from ken.Context fully
		logger := defaultLogger

		// Add user info from event if available
		if ctx != nil && ctx.GetEvent() != nil && ctx.GetEvent().Member != nil && ctx.GetEvent().Member.User != nil {
			logger = logger.With().
				Str("user", ctx.GetEvent().Member.User.Username).
				Str("user_id", ctx.GetEvent().Member.User.ID).
				Logger()
		}

		logger.Error().
			Any("panic", r).
			Bytes("stack", debug.Stack()).
			Msg("Component handler panicked")

		// Try to send error message to Discord
		_ = ctx.RespondError("A fatal error occurred in the component handler", "Error")
	}
}

// WrapKenComponent wraps a Ken component handler with panic recovery
func WrapKenComponent(handler func(ken.ComponentContext) bool) func(ken.ComponentContext) bool {
	return func(ctx ken.ComponentContext) bool {
		defer RecoverKenComponent(ctx)
		return handler(ctx)
	}
}
