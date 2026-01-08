package logger

import (
	"io"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// Config holds logger configuration
type Config struct {
	Environment string        // "local", "staging", "production"
	DBPool      *pgxpool.Pool // Database connection pool for logging
}

var (
	// defaultLogger is the global logger instance
	defaultLogger zerolog.Logger
	// defaultDBWriter holds the database writer instance
	defaultDBWriter *DatabaseWriter
)

// Init initializes the global logger based on environment
func Init(cfg Config) (zerolog.Logger, error) {
	env := strings.ToLower(cfg.Environment)
	if env == "" {
		env = "local"
	}

	var writers []io.Writer

	// Console output - always include
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
		NoColor:    true, // Disable color codes for better readability in DB/Railway
	}
	writers = append(writers, consoleWriter)

	// Database output - include if pool is provided
	if cfg.DBPool != nil {
		defaultDBWriter = NewDatabaseWriter(cfg.DBPool, env)
		if defaultDBWriter != nil {
			writers = append(writers, defaultDBWriter)
		}
	}

	// Determine log level based on environment
	logLevel := zerolog.DebugLevel
	if env == "production" || env == "staging" {
		logLevel = zerolog.InfoLevel
	}

	// Create multi-writer if we have multiple writers
	var writer io.Writer
	if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = io.MultiWriter(writers...)
	}

	// Create logger with timestamp and caller info
	defaultLogger = zerolog.New(writer).
		With().
		Timestamp().
		Logger().
		Level(logLevel)

	// Set as default
	zerolog.DefaultContextLogger = &defaultLogger

	return defaultLogger, nil
}

// Get returns a pointer to the global logger instance
func Get() *zerolog.Logger {
	return &defaultLogger
}

// SetLevel updates the global log level
func SetLevel(level zerolog.Level) {
	defaultLogger = defaultLogger.Level(level)
}

// Close performs cleanup (used during shutdown)
func Close() error {
	if defaultDBWriter != nil {
		return defaultDBWriter.Close()
	}
	return nil
}
