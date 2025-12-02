package logger

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// LogEntry represents a single log entry to be inserted into the database
type LogEntry struct {
	Timestamp     time.Time
	Level         string
	Message       string
	CorrelationID *string
	UserID        *int64
	CommandName   *string
	ErrorDetails  map[string]interface{}
	RequestData   map[string]interface{}
	Environment   string
}

// DatabaseWriter implements io.Writer and writes logs to PostgreSQL asynchronously
type DatabaseWriter struct {
	pool        *pgxpool.Pool
	channel     chan LogEntry
	done        chan struct{}
	wg          sync.WaitGroup
	batchSize   int
	flushTimer  *time.Ticker
	environment string
}

// NewDatabaseWriter creates a new database writer with async batching
func NewDatabaseWriter(pool *pgxpool.Pool, environment string) *DatabaseWriter {
	if pool == nil {
		return nil // Database writer is optional
	}

	dw := &DatabaseWriter{
		pool:        pool,
		channel:     make(chan LogEntry, 100), // Buffered channel to avoid blocking
		done:        make(chan struct{}),
		batchSize:   100,
		flushTimer:  time.NewTicker(5 * time.Second),
		environment: environment,
	}

	// Start background worker goroutine
	dw.wg.Add(1)
	go dw.batchWorker()

	return dw
}

// Write parses a zerolog JSON line and enqueues it for database insertion
func (dw *DatabaseWriter) Write(p []byte) (n int, err error) {
	if dw == nil || dw.pool == nil {
		return len(p), nil
	}

	// Parse zerolog JSON output
	var entry map[string]interface{}
	if err := json.Unmarshal(p, &entry); err != nil {
		return len(p), nil // Silently skip malformed entries
	}

	logEntry := LogEntry{
		Environment:  dw.environment,
		ErrorDetails: make(map[string]interface{}),
		RequestData:  make(map[string]interface{}),
	}

	// Extract standard fields
	if ts, ok := entry["time"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			logEntry.Timestamp = t
		}
	} else {
		logEntry.Timestamp = time.Now()
	}

	if level, ok := entry["level"].(string); ok {
		logEntry.Level = level
	}

	if msg, ok := entry["message"].(string); ok {
		logEntry.Message = msg
	}

	// Extract structured fields
	for key, value := range entry {
		switch key {
		case "time", "level", "message":
			continue
		case "correlation_id":
			if str, ok := value.(string); ok {
				logEntry.CorrelationID = &str
			}
		case "user_id":
			if str, ok := value.(string); ok {
				// Try to convert to int64 (Discord IDs are large)
				var id int64
				json.Unmarshal([]byte(str), &id)
				if id > 0 {
					logEntry.UserID = &id
				}
			}
		case "command":
			if str, ok := value.(string); ok {
				logEntry.CommandName = &str
			}
		case "user":
			logEntry.RequestData["user"] = value
		case "error":
			logEntry.ErrorDetails["error"] = value
		default:
			// Store other fields in request_data
			logEntry.RequestData[key] = value
		}
	}

	// Non-blocking send to channel
	select {
	case dw.channel <- logEntry:
	case <-dw.done:
		return len(p), nil
	default:
		// Channel full, skip to avoid blocking
	}

	return len(p), nil
}

// Sync is a no-op for compatibility with io.Writer
func (dw *DatabaseWriter) Sync() error {
	return nil
}

// batchWorker accumulates log entries and inserts them in batches
func (dw *DatabaseWriter) batchWorker() {
	defer dw.wg.Done()

	batch := make([]LogEntry, 0, dw.batchSize)

	for {
		select {
		case entry := <-dw.channel:
			batch = append(batch, entry)
			if len(batch) >= dw.batchSize {
				dw.insertBatch(batch)
				batch = batch[:0]
			}

		case <-dw.flushTimer.C:
			if len(batch) > 0 {
				dw.insertBatch(batch)
				batch = batch[:0]
			}

		case <-dw.done:
			// Drain remaining entries
			close(dw.channel)
			for entry := range dw.channel {
				batch = append(batch, entry)
			}
			if len(batch) > 0 {
				dw.insertBatch(batch)
			}
			dw.flushTimer.Stop()
			return
		}
	}
}

// insertBatch performs a batch insert of log entries into the database
func (dw *DatabaseWriter) insertBatch(batch []LogEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, entry := range batch {
		errorJSON, _ := json.Marshal(entry.ErrorDetails)
		requestJSON, _ := json.Marshal(entry.RequestData)

		query := `
			INSERT INTO logs (
				timestamp, level, message, correlation_id, user_id,
				command_name, error_details, request_data, environment
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		if err := dw.pool.QueryRow(ctx, query,
			entry.Timestamp,
			entry.Level,
			entry.Message,
			entry.CorrelationID,
			entry.UserID,
			entry.CommandName,
			errorJSON,
			requestJSON,
			entry.Environment,
		).Scan(); err != nil && err.Error() != "no rows in result set" {
			// Log to stderr on failure (avoid infinite loop)
			zerolog.DefaultContextLogger.Error().Err(err).Msg("Failed to insert log batch")
		}
	}
}

// Close gracefully shuts down the database writer and flushes pending logs
func (dw *DatabaseWriter) Close() error {
	if dw == nil || dw.pool == nil {
		return nil
	}

	close(dw.done)
	dw.wg.Wait()
	return nil
}
