package logger

import (
	"runtime/debug"

	"github.com/rs/zerolog"
)

// SafeGo launches a goroutine with panic recovery and error logging
func SafeGo(logger zerolog.Logger, name string, fn func() error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error().
					Any("panic", r).
					Bytes("stack", debug.Stack()).
					Str("goroutine", name).
					Msg("Goroutine panicked")
			}
		}()

		if err := fn(); err != nil {
			logger.Error().
				Err(err).
				Str("goroutine", name).
				Msg("Goroutine failed")
		}
	}()
}

// SafeGoVoid launches a goroutine without error return, only for cleanup operations
func SafeGoVoid(logger zerolog.Logger, name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error().
					Any("panic", r).
					Bytes("stack", debug.Stack()).
					Str("goroutine", name).
					Msg("Goroutine panicked")
			}
		}()

		fn()
	}()
}
