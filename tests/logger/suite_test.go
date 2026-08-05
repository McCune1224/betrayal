package logger

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/suite"
)

type LoggerTestSuite struct {
	suite.Suite
	DB *pgxpool.Pool
}

func (lts *LoggerTestSuite) SetupSuite() {
	lts.DB = testutil.NewTestPool(lts.T())
}

func (lts *LoggerTestSuite) SetupTest() {
	// Clean up logs table before each test.
	_, err := lts.DB.Exec(context.Background(), "TRUNCATE TABLE logs CASCADE")
	if err != nil {
		lts.FailNow(err.Error())
	}
}

func (lts *LoggerTestSuite) TearDownSuite() {
	if lts.DB != nil {
		lts.DB.Close()
	}
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

// TestMain boots the suite: loads env, enforces the production guard,
// serializes against other DB suites, and applies migrations once.
func TestMain(m *testing.M) {
	os.Exit(testutil.Bootstrap(m))
}
