package logger

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type LoggerTestSuite struct {
	suite.Suite
	DB *pgxpool.Pool
}

func (lts *LoggerTestSuite) SetupTest() {
	godotenv.Load(".env")
	godotenv.Load("../.env")
	godotenv.Load("../../.env")
	pools, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		lts.FailNow(err.Error())
	}
	lts.DB = pools

	// Clean up logs table before each test
	_, err = lts.DB.Exec(context.Background(), "TRUNCATE TABLE logs CASCADE")
	if err != nil {
		lts.FailNow(err.Error())
	}
}

func (lts *LoggerTestSuite) TearDownTest() {
	if lts.DB != nil {
		lts.DB.Close()
	}
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}
