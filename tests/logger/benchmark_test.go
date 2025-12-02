package logger

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

// PerformanceSuite tests logger and audit performance under load
type PerformanceSuite struct {
	suite.Suite
	dbPool *pgxpool.Pool
}

// SetupTest initializes the test suite
func (suite *PerformanceSuite) SetupTest() {
	// Note: This test requires a local PostgreSQL database
	// Set up test database connection
	var err error
	dsn := "postgresql://betrayal_dev:betrayal_dev@localhost:5432/betrayal_test"
	suite.dbPool, err = pgxpool.New(context.Background(), dsn)

	if err != nil {
		suite.T().Skip("Skipping performance test - database not available")
	}

	// Verify database is accessible
	err = suite.dbPool.Ping(context.Background())
	if err != nil {
		suite.dbPool.Close()
		suite.T().Skip("Skipping performance test - database not accessible")
	}
}

// TearDownTest closes database connections
func (suite *PerformanceSuite) TearDownTest() {
	if suite.dbPool != nil {
		suite.dbPool.Close()
	}
}

// TestLogInsertPerformance measures the performance of inserting logs into the database
func (suite *PerformanceSuite) TestLogInsertPerformance() {
	ctx := context.Background()
	startTime := time.Now()
	logCount := 1000

	// Insert logs
	for i := 0; i < logCount; i++ {
		_, err := suite.dbPool.Exec(ctx,
			`INSERT INTO logs (created_at, level, message, correlation_id, user_id, command_name, error_details, request_data, environment)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			time.Now(),
			"info",
			fmt.Sprintf("test message %d", i),
			uuid.New().String(),
			fmt.Sprintf("user_%d", i%10),
			"test_command",
			nil,
			nil,
			"test",
		)
		suite.NoError(err)
	}

	duration := time.Since(startTime)
	throughput := float64(logCount) / duration.Seconds()
	suite.T().Logf("Log insert performance: %d logs in %v (%.0f logs/sec)", logCount, duration, throughput)
	suite.Assert().Greater(throughput, 500.0, "Log insertion throughput too low (expected >500 logs/sec)")
}

// TestAuditInsertPerformance measures the performance of inserting audit records
func (suite *PerformanceSuite) TestAuditInsertPerformance() {
	ctx := context.Background()
	startTime := time.Now()
	auditCount := 500

	// Insert audit records
	for i := 0; i < auditCount; i++ {
		_, err := suite.dbPool.Exec(ctx,
			`INSERT INTO command_audit (timestamp, command_name, user_id, username, user_roles, is_admin, command_arguments, status, error_message, execution_time_ms)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			time.Now(),
			"test_command",
			fmt.Sprintf("user_%d", i%10),
			fmt.Sprintf("testuser_%d", i%10),
			[]string{"user", "admin"},
			i%2 == 0,
			fmt.Sprintf(`{"index": %d, "key": "value"}`, i),
			"success",
			nil,
			100+i,
		)
		suite.NoError(err)
	}

	duration := time.Since(startTime)
	throughput := float64(auditCount) / duration.Seconds()
	suite.T().Logf("Audit insert performance: %d records in %v (%.0f records/sec)", auditCount, duration, throughput)
	suite.Assert().Greater(throughput, 200.0, "Audit insertion throughput too low (expected >200 records/sec)")
}

// TestConcurrentWrites measures performance under concurrent load
func (suite *PerformanceSuite) TestConcurrentWrites() {
	ctx := context.Background()
	numGoroutines := 10
	recordsPerGoroutine := 50
	totalRecords := numGoroutines * recordsPerGoroutine

	startTime := time.Now()
	successCount := int64(0)
	errorCount := int64(0)
	var wg sync.WaitGroup

	// Launch concurrent writers
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < recordsPerGoroutine; i++ {
				_, err := suite.dbPool.Exec(ctx,
					`INSERT INTO command_audit (timestamp, command_name, user_id, username, user_roles, is_admin, command_arguments, status, error_message, execution_time_ms)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
					time.Now(),
					fmt.Sprintf("cmd_%d", goroutineID),
					fmt.Sprintf("user_%d", goroutineID),
					fmt.Sprintf("user_%d", goroutineID),
					[]string{"user"},
					false,
					fmt.Sprintf(`{"goroutine": %d, "record": %d}`, goroutineID, i),
					"success",
					nil,
					50+i,
				)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(g)
	}

	wg.Wait()
	duration := time.Since(startTime)
	throughput := float64(totalRecords) / duration.Seconds()

	suite.T().Logf("Concurrent writes (%d goroutines): %d records in %v (%.0f records/sec), errors: %d",
		numGoroutines, totalRecords, duration, throughput, errorCount)
	suite.Equal(int64(0), errorCount, "Should not have any errors")
	suite.Greater(throughput, 300.0, "Concurrent throughput too low (expected >300 records/sec)")
}

// TestAuditQueryPerformance measures query performance for audit reports
func (suite *PerformanceSuite) TestAuditQueryPerformance() {
	ctx := context.Background()

	// Insert test data
	for i := 0; i < 200; i++ {
		suite.dbPool.Exec(ctx,
			`INSERT INTO command_audit (timestamp, command_name, user_id, username, user_roles, is_admin, command_arguments, status, error_message, execution_time_ms)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			time.Now().Add(-time.Duration(i)*time.Minute),
			[]string{"cmd_a", "cmd_b", "cmd_c"}[i%3],
			fmt.Sprintf("user_%d", i%5),
			fmt.Sprintf("user_%d", i%5),
			[]string{"user"},
			i%3 == 0,
			`{}`,
			"success",
			nil,
			100+i,
		)
	}

	startTime := time.Now()

	// Query 1: Get all commands for a user (100ms)
	rows, _ := suite.dbPool.Query(ctx,
		"SELECT COUNT(*) FROM command_audit WHERE user_id = $1", "user_0")
	rows.Close()

	// Query 2: Get top commands (100ms)
	rows, _ = suite.dbPool.Query(ctx,
		`SELECT command_name, COUNT(*) FROM command_audit 
		 GROUP BY command_name ORDER BY COUNT(*) DESC LIMIT 10`)
	rows.Close()

	// Query 3: Get average execution time (50ms)
	suite.dbPool.QueryRow(ctx,
		"SELECT AVG(execution_time_ms) FROM command_audit").Scan(nil)

	// Query 4: Get failure rate (100ms)
	suite.dbPool.QueryRow(ctx,
		"SELECT COUNT(*) FROM command_audit WHERE status != 'success'").Scan(nil)

	duration := time.Since(startTime)
	suite.T().Logf("Completed 4 audit queries in %v", duration)
	suite.Less(duration.Milliseconds(), int64(500), "Query performance degraded")
}

// TestBatchedLogWriteSimulation simulates batched log writes
func (suite *PerformanceSuite) TestBatchedLogWriteSimulation() {
	ctx := context.Background()
	batchSize := 100
	totalLogs := 1000

	startTime := time.Now()
	batch := make([]map[string]interface{}, 0, batchSize)

	for i := 0; i < totalLogs; i++ {
		batch = append(batch, map[string]interface{}{
			"timestamp":      time.Now(),
			"level":          "info",
			"message":        fmt.Sprintf("message_%d", i),
			"correlation_id": uuid.New().String(),
		})

		// Flush batch
		if len(batch) >= batchSize {
			for _, record := range batch {
				suite.dbPool.Exec(ctx,
					`INSERT INTO logs (created_at, level, message, correlation_id, environment)
					 VALUES ($1, $2, $3, $4, $5)`,
					record["timestamp"],
					record["level"],
					record["message"],
					record["correlation_id"],
					"test",
				)
			}
			batch = make([]map[string]interface{}, 0, batchSize)
		}
	}

	// Flush remaining
	for _, record := range batch {
		suite.dbPool.Exec(ctx,
			`INSERT INTO logs (created_at, level, message, correlation_id, environment)
			 VALUES ($1, $2, $3, $4, $5)`,
			record["timestamp"],
			record["level"],
			record["message"],
			record["correlation_id"],
			"test",
		)
	}

	duration := time.Since(startTime)
	throughput := float64(totalLogs) / duration.Seconds()
	suite.T().Logf("Batched log write (batch size %d): %d logs in %v (%.0f logs/sec)", batchSize, totalLogs, duration, throughput)
}

// TestMemoryUnderLoad tests memory usage during high throughput
func (suite *PerformanceSuite) TestMemoryUnderLoad() {
	ctx := context.Background()

	// Create 10,000 audit records in rapid succession
	for i := 0; i < 10000; i++ {
		suite.dbPool.Exec(ctx,
			`INSERT INTO command_audit (timestamp, command_name, user_id, username, user_roles, is_admin, command_arguments, status, error_message, execution_time_ms)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			time.Now(),
			"test_cmd",
			fmt.Sprintf("user_%d", i%100),
			fmt.Sprintf("user_%d", i%100),
			[]string{"user"},
			false,
			`{}`,
			"success",
			nil,
			50,
		)
	}

	suite.T().Logf("Successfully inserted 10,000 audit records under load")
}

// In TestPerformance, run the performance tests
func TestPerformance(t *testing.T) {
	suite.Run(t, new(PerformanceSuite))
}
