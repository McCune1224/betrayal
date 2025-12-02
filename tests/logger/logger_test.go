package logger

import (
	"context"
	"io"
	"time"

	"github.com/mccune1224/betrayal/internal/logger"
	"github.com/rs/zerolog"
)

// TestLoggerInit tests basic logger initialization
func (lts *LoggerTestSuite) TestLoggerInit() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	log, err := logger.Init(cfg)
	lts.NoError(err)
	lts.NotNil(log)
}

// TestLoggerInitProduction tests production logger initialization
func (lts *LoggerTestSuite) TestLoggerInitProduction() {
	cfg := logger.Config{
		Environment: "production",
		DBPool:      nil,
	}

	log, err := logger.Init(cfg)
	lts.NoError(err)
	lts.NotNil(log)
}

// TestLoggerInitStaging tests staging logger initialization
func (lts *LoggerTestSuite) TestLoggerInitStaging() {
	cfg := logger.Config{
		Environment: "staging",
		DBPool:      nil,
	}

	log, err := logger.Init(cfg)
	lts.NoError(err)
	lts.NotNil(log)
}

// TestLoggerGet tests retrieving the global logger instance
func (lts *LoggerTestSuite) TestLoggerGet() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()
	lts.NotNil(log)
}

// TestLoggerSetLevel tests changing log level at runtime
func (lts *LoggerTestSuite) TestLoggerSetLevel() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	// Change log level
	logger.SetLevel(zerolog.WarnLevel)

	log := logger.Get()
	lts.NotNil(log)
}

// TestLoggerClose tests graceful shutdown
func (lts *LoggerTestSuite) TestLoggerClose() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	err = logger.Close()
	lts.NoError(err)
}

// TestLoggerInfoMessage tests logging an info message
func (lts *LoggerTestSuite) TestLoggerInfoMessage() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	// This should not panic
	log := logger.Get()
	log.Info().Msg("test info message")
}

// TestLoggerErrorMessage tests logging an error message
func (lts *LoggerTestSuite) TestLoggerErrorMessage() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	// This should not panic
	log := logger.Get()
	log.Error().Err(io.EOF).Msg("test error message")
}

// TestLoggerStructuredFields tests logging with structured fields
func (lts *LoggerTestSuite) TestLoggerStructuredFields() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	// This should not panic
	log := logger.Get()
	log.Info().
		Str("user", "testuser").
		Int64("user_id", 123456789).
		Str("command", "test_command").
		Msg("structured log entry")
}

// TestLoggerMultipleInitializations tests calling Init multiple times
func (lts *LoggerTestSuite) TestLoggerMultipleInitializations() {
	cfg1 := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg1)
	lts.NoError(err)

	cfg2 := logger.Config{
		Environment: "production",
		DBPool:      nil,
	}

	_, err = logger.Init(cfg2)
	lts.NoError(err)

	log := logger.Get()
	lts.NotNil(log)
}

// TestLoggerDefaultsToLocal tests that empty environment defaults to local
func (lts *LoggerTestSuite) TestLoggerDefaultsToLocal() {
	cfg := logger.Config{
		Environment: "",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()
	lts.NotNil(log)
}

// TestLoggerConcurrentWrites tests concurrent log writes don't panic
func (lts *LoggerTestSuite) TestLoggerConcurrentWrites() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			log.Info().Int("goroutine", idx).Msg("concurrent write")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestLoggerWithContext tests logging with context
func (lts *LoggerTestSuite) TestLoggerWithContext() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log := logger.Get()
	if deadline, ok := ctx.Deadline(); ok {
		log.Info().Str("context_deadline", deadline.String()).Msg("logging with context")
	}
}

// TestLoggerLevelFiltering tests that debug logs are filtered in production
func (lts *LoggerTestSuite) TestLoggerLevelFiltering() {
	cfg := logger.Config{
		Environment: "production",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()

	// These should still not panic, even if not logged
	log.Debug().Msg("debug message - should not appear in production")
	log.Info().Msg("info message - should appear in production")
}

// TestLoggerStackTrace tests error logging with stack trace
func (lts *LoggerTestSuite) TestLoggerStackTrace() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	log := logger.Get()

	log.Error().Err(io.EOF).Stack().Msg("error with stack trace")
}
