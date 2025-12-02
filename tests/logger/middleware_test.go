package logger

import (
	"github.com/google/uuid"
	"github.com/mccune1224/betrayal/internal/logger"
)

// TestGenerateCorrelationID tests UUID generation
func (lts *LoggerTestSuite) TestGenerateCorrelationID() {
	id := logger.GenerateCorrelationID()

	lts.NotNil(id)
	lts.NotEqual(uuid.UUID{}, id)
}

// TestGenerateCorrelationIDUniqueness tests that each call generates a unique ID
func (lts *LoggerTestSuite) TestGenerateCorrelationIDUniqueness() {
	id1 := logger.GenerateCorrelationID()
	id2 := logger.GenerateCorrelationID()

	lts.NotEqual(id1, id2)
}

// TestGenerateCorrelationIDMultiple tests generating multiple IDs
func (lts *LoggerTestSuite) TestGenerateCorrelationIDMultiple() {
	ids := make(map[uuid.UUID]bool)

	for i := 0; i < 100; i++ {
		id := logger.GenerateCorrelationID()
		lts.False(ids[id], "Generated duplicate correlation ID")
		ids[id] = true
	}

	lts.Equal(100, len(ids))
}

// TestCorrelationIDString tests that correlation ID can be stringified
func (lts *LoggerTestSuite) TestCorrelationIDString() {
	id := logger.GenerateCorrelationID()
	idStr := id.String()

	lts.NotEmpty(idStr)
	lts.Len(idStr, 36) // Standard UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
}

// TestFromKenContextNil tests handling of nil Ken context
func (lts *LoggerTestSuite) TestFromKenContextNil() {
	cfg := logger.Config{
		Environment: "local",
		DBPool:      nil,
	}

	_, err := logger.Init(cfg)
	lts.NoError(err)

	// This should not panic with nil context
	defer func() {
		if r := recover(); r != nil {
			lts.Fail("panic occurred with nil context")
		}
	}()

	log := logger.FromKenContext(nil)
	lts.NotNil(log)
}

// TestInjectKenContextNil tests handling of nil Ken context injection
func (lts *LoggerTestSuite) TestInjectKenContextNil() {
	// This should not panic
	logger.InjectKenContext(nil)
}
