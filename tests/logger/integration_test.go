//go:build integration

package logger

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"github.com/mccune1224/betrayal/internal/logger"
)

// AuditIntegrationSuite tests the full audit flow with multiple commands
type AuditIntegrationSuite struct {
	suite.Suite
	dbPool *pgxpool.Pool
	writer *logger.AuditWriter
}

// SetupTest initializes the integration test suite
func (suite *AuditIntegrationSuite) SetupTest() {
	var err error
	dsn := "postgresql://betrayal_dev:betrayal_dev@localhost:5432/betrayal_test"
	suite.dbPool, err = pgxpool.New(context.Background(), dsn)

	if err != nil {
		suite.T().Skip("Skipping integration test - database not available")
	}

	// Verify database is accessible
	err = suite.dbPool.Ping(context.Background())
	if err != nil {
		suite.dbPool.Close()
		suite.T().Skip("Skipping integration test - database not accessible")
	}

	// Initialize audit writer
	suite.writer = logger.NewAuditWriter(suite.dbPool, "test")
}

// TearDownTest closes connections and cleanup
func (suite *AuditIntegrationSuite) TearDownTest() {
	if suite.writer != nil {
		suite.writer.Close()
	}
	if suite.dbPool != nil {
		// Clean up test data
		suite.dbPool.Exec(context.Background(), "DELETE FROM command_audit WHERE user_id LIKE 'test_%'")
		suite.dbPool.Close()
	}
}

// TestSingleCommandAudit tests audit logging for a single command
func (suite *AuditIntegrationSuite) TestSingleCommandAudit() {
	audit := logger.CommandAudit{
		CorrelationID:    "test-corr-id-1",
		CommandName:      "test_command",
		UserID:           "test_user_1",
		Username:         "testuser1",
		UserRoles:        []string{"user", "admin"},
		GuildID:          "guild_123",
		ChannelID:        "channel_123",
		IsAdmin:          true,
		CommandArguments: map[string]interface{}{"option1": "value1"},
		Status:           "success",
		ExecutionTimeMs:  100,
	}

	// Log the command
	suite.writer.LogCommand(audit)

	// Wait for batch to flush
	time.Sleep(100 * time.Millisecond)

	// Verify it was logged
	var count int64
	err := suite.dbPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM command_audit WHERE correlation_id = $1",
		"test-corr-id-1").Scan(&count)

	suite.NoError(err)
	suite.Equal(int64(1), count, "Audit record should be logged")
}

// TestMultipleCommandsAudit tests audit logging for multiple commands
func (suite *AuditIntegrationSuite) TestMultipleCommandsAudit() {
	numCommands := 100

	// Log multiple commands
	for i := 0; i < numCommands; i++ {
		audit := logger.CommandAudit{
			CorrelationID:    fmt.Sprintf("test-corr-id-%d", i),
			CommandName:      fmt.Sprintf("cmd_%d", i%5),
			UserID:           fmt.Sprintf("test_user_%d", i%10),
			Username:         fmt.Sprintf("testuser_%d", i%10),
			UserRoles:        []string{"user"},
			GuildID:          "guild_123",
			ChannelID:        "channel_123",
			IsAdmin:          i%3 == 0,
			CommandArguments: map[string]interface{}{"index": i},
			Status:           "success",
			ExecutionTimeMs:  int32(50 + i%50),
		}
		suite.writer.LogCommand(audit)
	}

	// Wait for all batches to flush
	time.Sleep(500 * time.Millisecond)

	// Verify all records were logged
	var count int64
	err := suite.dbPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM command_audit WHERE user_id LIKE 'test_user_%'").Scan(&count)

	suite.NoError(err)
	suite.Equal(int64(numCommands), count, "All audit records should be logged")
}

// TestErrorCommandAudit tests audit logging for failed commands
func (suite *AuditIntegrationSuite) TestErrorCommandAudit() {
	errorMsg := "command execution failed"
	audit := logger.CommandAudit{
		CorrelationID:    "test-corr-error-1",
		CommandName:      "failed_command",
		UserID:           "test_user_error",
		Username:         "testuser_error",
		UserRoles:        []string{"user"},
		GuildID:          "guild_123",
		ChannelID:        "channel_123",
		IsAdmin:          false,
		CommandArguments: map[string]interface{}{},
		Status:           "error",
		ErrorMessage:     &errorMsg,
		ExecutionTimeMs:  50,
	}

	suite.writer.LogCommand(audit)
	time.Sleep(100 * time.Millisecond)

	// Verify error record
	var retrievedErr *string
	err := suite.dbPool.QueryRow(context.Background(),
		"SELECT error_message FROM command_audit WHERE correlation_id = $1",
		"test-corr-error-1").Scan(&retrievedErr)

	suite.NoError(err)
	suite.NotNil(retrievedErr, "Error message should be stored")
	suite.Equal(errorMsg, *retrievedErr, "Error message should match")
}

// TestConcurrentCommandAudit tests concurrent audit logging
func (suite *AuditIntegrationSuite) TestConcurrentCommandAudit() {
	numGoroutines := 10
	commandsPerGoroutine := 20
	totalCommands := numGoroutines * commandsPerGoroutine

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < commandsPerGoroutine; i++ {
				audit := logger.CommandAudit{
					CorrelationID:    fmt.Sprintf("concurrent-%d-%d", goroutineID, i),
					CommandName:      fmt.Sprintf("cmd_%d", goroutineID),
					UserID:           fmt.Sprintf("test_user_%d", goroutineID),
					Username:         fmt.Sprintf("testuser_%d", goroutineID),
					UserRoles:        []string{"user"},
					GuildID:          "guild_123",
					ChannelID:        "channel_123",
					IsAdmin:          false,
					CommandArguments: map[string]interface{}{"goroutine": goroutineID},
					Status:           "success",
					ExecutionTimeMs:  int32(50 + i),
				}
				suite.writer.LogCommand(audit)
			}
		}(g)
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	// Verify all concurrent commands were logged
	var count int64
	err := suite.dbPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM command_audit WHERE correlation_id LIKE 'concurrent-%'").Scan(&count)

	suite.NoError(err)
	suite.Equal(int64(totalCommands), count, "All concurrent audit records should be logged")
}

// TestAuditDataIntegrity tests that audit data is preserved correctly
func (suite *AuditIntegrationSuite) TestAuditDataIntegrity() {
	originalArgs := map[string]interface{}{
		"user_mentions": "user_123,user_456",
		"channel":       "general",
		"options":       map[string]interface{}{"nested": "value"},
	}

	audit := logger.CommandAudit{
		CorrelationID:    "test-integrity-1",
		CommandName:      "data_integrity_cmd",
		UserID:           "test_user_integrity",
		Username:         "testuser_integrity",
		UserRoles:        []string{"user", "moderator", "admin"},
		GuildID:          "guild_999",
		ChannelID:        "channel_999",
		IsAdmin:          true,
		CommandArguments: originalArgs,
		Status:           "success",
		ExecutionTimeMs:  250,
	}

	suite.writer.LogCommand(audit)
	time.Sleep(100 * time.Millisecond)

	// Verify all data was preserved
	var cmdName, userID, username string
	var isAdmin bool
	var execTime int32
	err := suite.dbPool.QueryRow(context.Background(),
		`SELECT command_name, user_id, username, is_admin, execution_time_ms 
		 FROM command_audit WHERE correlation_id = $1`,
		"test-integrity-1").Scan(&cmdName, &userID, &username, &isAdmin, &execTime)

	suite.NoError(err)
	suite.Equal("data_integrity_cmd", cmdName)
	suite.Equal("test_user_integrity", userID)
	suite.Equal("testuser_integrity", username)
	suite.True(isAdmin)
	suite.Equal(int32(250), execTime)
}

// TestAuditStatistics tests querying audit statistics
func (suite *AuditIntegrationSuite) TestAuditStatistics() {
	// Insert diverse audit data
	commands := []string{"cmd_a", "cmd_b", "cmd_c"}
	users := []string{"user_1", "user_2", "user_3"}

	for i := 0; i < 30; i++ {
		audit := logger.CommandAudit{
			CorrelationID:    fmt.Sprintf("stats-test-%d", i),
			CommandName:      commands[i%3],
			UserID:           users[i%3],
			Username:         fmt.Sprintf("user_%d", i%3),
			UserRoles:        []string{"user"},
			GuildID:          "guild_123",
			ChannelID:        "channel_123",
			IsAdmin:          false,
			CommandArguments: map[string]interface{}{},
			Status: func() string {
				if i%5 == 0 {
					return "error"
				} else {
					return "success"
				}
			}(),
			ExecutionTimeMs: int32(50 + i),
		}
		suite.writer.LogCommand(audit)
	}

	time.Sleep(500 * time.Millisecond)

	// Test: Count total commands from stats tests
	var total int64
	err := suite.dbPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM command_audit WHERE correlation_id LIKE 'stats-test-%'").Scan(&total)
	suite.NoError(err)
	suite.Equal(int64(30), total)

	// Test: Count failed commands
	var failures int64
	err = suite.dbPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM command_audit WHERE correlation_id LIKE 'stats-test-%' AND status = 'error'").Scan(&failures)
	suite.NoError(err)
	suite.Equal(int64(6), failures) // 30 / 5 = 6 failures

	// Test: Get average execution time
	var avgExecTime float64
	err = suite.dbPool.QueryRow(context.Background(),
		"SELECT AVG(execution_time_ms) FROM command_audit WHERE correlation_id LIKE 'stats-test-%'").Scan(&avgExecTime)
	suite.NoError(err)
	suite.Greater(avgExecTime, 50.0)
}

// TestAuditBatchFlushing tests that batches are flushed correctly
func (suite *AuditIntegrationSuite) TestAuditBatchFlushing() {
	// Write exactly one batch (50 records)
	for i := 0; i < 50; i++ {
		audit := logger.CommandAudit{
			CorrelationID:    fmt.Sprintf("batch-test-1-%d", i),
			CommandName:      "batch_cmd",
			UserID:           "test_user_batch",
			Username:         "testuser_batch",
			UserRoles:        []string{"user"},
			GuildID:          "guild_123",
			ChannelID:        "channel_123",
			IsAdmin:          false,
			CommandArguments: map[string]interface{}{},
			Status:           "success",
			ExecutionTimeMs:  100,
		}
		suite.writer.LogCommand(audit)
	}

	// Wait for batch flush (should flush immediately at 50 records)
	time.Sleep(100 * time.Millisecond)

	var count int64
	err := suite.dbPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM command_audit WHERE correlation_id LIKE 'batch-test-1-%'").Scan(&count)
	suite.NoError(err)
	suite.Equal(int64(50), count, "Batch should be flushed when reaching batch size")
}

func TestAuditIntegration(t *testing.T) {
	suite.Run(t, new(AuditIntegrationSuite))
}
