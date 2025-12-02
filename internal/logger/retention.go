package logger

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// RetentionConfig holds log retention settings
type RetentionConfig struct {
	RetentionDays int    // Days to keep logs (90)
	ArchiveDir    string // Directory to archive old logs
}

// StartRetentionWorker starts a background goroutine that manages log retention
// It runs once daily at midnight to clean and archive old logs
func StartRetentionWorker(pool *pgxpool.Pool, logger zerolog.Logger, cfg RetentionConfig) {
	if cfg.RetentionDays <= 0 {
		return // Retention disabled
	}

	SafeGo(logger, "log_retention", func() error {
		for {
			now := time.Now()
			// Schedule for next midnight
			nextMidnight := now.AddDate(0, 0, 1)
			nextMidnight = time.Date(nextMidnight.Year(), nextMidnight.Month(), nextMidnight.Day(), 0, 0, 0, 0, nextMidnight.Location())

			waitDuration := nextMidnight.Sub(now)
			logger.Info().
				Str("next_run", nextMidnight.String()).
				Msg("Log retention scheduled")

			select {
			case <-time.After(waitDuration):
				// Run archival and cleanup
				err := archiveAndCleanLogs(pool, logger, cfg)
				if err != nil {
					logger.Error().Err(err).Msg("Log retention job failed")
				} else {
					logger.Info().Msg("Log retention job completed successfully")
				}
			}
		}
	})
}

// archiveAndCleanLogs exports old logs to CSV and deletes them from database
func archiveAndCleanLogs(pool *pgxpool.Pool, logger zerolog.Logger, cfg RetentionConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoffTime := time.Now().AddDate(0, 0, -cfg.RetentionDays)

	// Create archive directory if it doesn't exist
	if cfg.ArchiveDir != "" {
		if err := os.MkdirAll(cfg.ArchiveDir, 0755); err != nil {
			return fmt.Errorf("failed to create archive directory: %w", err)
		}
	}

	// Export logs older than retention period to CSV
	if err := exportLogsToCSV(ctx, pool, logger, cutoffTime, cfg.ArchiveDir); err != nil {
		logger.Error().Err(err).Msg("Failed to export logs to CSV")
		return err
	}

	// Delete archived logs from database
	query := `DELETE FROM logs WHERE created_at < $1`
	result, err := pool.Exec(ctx, query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to delete old logs: %w", err)
	}

	rowsDeleted := result.RowsAffected()
	logger.Info().
		Int64("rows_deleted", rowsDeleted).
		Time("cutoff_time", cutoffTime).
		Msg("Old logs deleted from database")

	return nil
}

// exportLogsToCSV exports logs to a CSV file
func exportLogsToCSV(ctx context.Context, pool *pgxpool.Pool, logger zerolog.Logger, cutoffTime time.Time, archiveDir string) error {
	// Generate archive filename with date
	archiveFile := filepath.Join(archiveDir, fmt.Sprintf("logs_archive_%s.csv", cutoffTime.Format("2006-01-02")))

	file, err := os.Create(archiveFile)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	if err := writer.Write([]string{
		"timestamp", "level", "message", "correlation_id", "user_id",
		"command_name", "error_details", "request_data", "environment",
	}); err != nil {
		return err
	}

	// Query logs to archive
	query := `
		SELECT timestamp, level, message, correlation_id, user_id,
		       command_name, error_details, request_data, environment
		FROM logs
		WHERE created_at < $1
		ORDER BY timestamp DESC
		LIMIT 100000
	`

	rows, err := pool.Query(ctx, query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		var timestamp time.Time
		var level, message, environment string
		var correlationID, userID, commandName, errorDetails, requestData interface{}

		if err := rows.Scan(&timestamp, &level, &message, &correlationID, &userID,
			&commandName, &errorDetails, &requestData, &environment); err != nil {
			logger.Error().Err(err).Msg("Failed to scan log row")
			continue
		}

		// Convert nulls to empty strings for CSV
		corrIDStr := ""
		if correlationID != nil {
			corrIDStr = correlationID.(string)
		}
		userIDStr := ""
		if userID != nil {
			userIDStr = fmt.Sprintf("%v", userID)
		}
		cmdStr := ""
		if commandName != nil {
			cmdStr = commandName.(string)
		}
		errStr := ""
		if errorDetails != nil {
			errStr = errorDetails.(string)
		}
		reqStr := ""
		if requestData != nil {
			reqStr = requestData.(string)
		}

		if err := writer.Write([]string{
			timestamp.String(),
			level,
			message,
			corrIDStr,
			userIDStr,
			cmdStr,
			errStr,
			reqStr,
			environment,
		}); err != nil {
			return err
		}

		rowCount++
	}

	if err := rows.Err(); err != nil {
		return err
	}

	logger.Info().
		Str("archive_file", archiveFile).
		Int("row_count", rowCount).
		Msg("Logs exported to CSV")

	return nil
}
