package database

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	DB *pgxpool.Pool
	Q  *models.Queries
}

func (s *ConfigTestSuite) SetupTest() {
	godotenv.Load(".env")
	godotenv.Load("../.env")
	godotenv.Load("../../.env")

	pools, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	s.Require().NoError(err, "failed to connect to local test database (DATABASE_URL)")
	s.DB = pools
	s.Q = models.New(pools)
}

func (s *ConfigTestSuite) TearDownTest() {
	if s.DB != nil {
		s.DB.Close()
	}
}

// TestCommandLogChannel covers the singleton command-log channel row backing
// the logHandler lookup and the /channel log subcommand.
func (s *ConfigTestSuite) TestCommandLogChannel() {
	ctx := context.Background()

	// Clean slate.
	s.Require().NoError(s.Q.DeleteCommandLogChannel(ctx))

	// Not configured yet -> ErrNoRows.
	_, err := s.Q.GetCommandLogChannel(ctx)
	s.ErrorIs(err, pgx.ErrNoRows)

	// Set, read back.
	channelID, err := s.Q.SetCommandLogChannel(ctx, "111111111111111111")
	s.Require().NoError(err)
	s.Equal("111111111111111111", channelID)

	got, err := s.Q.GetCommandLogChannel(ctx)
	s.Require().NoError(err)
	s.Equal("111111111111111111", got)

	// Upsert replaces the single row.
	channelID, err = s.Q.SetCommandLogChannel(ctx, "222222222222222222")
	s.Require().NoError(err)
	s.Equal("222222222222222222", channelID)

	got, err = s.Q.GetCommandLogChannel(ctx)
	s.Require().NoError(err)
	s.Equal("222222222222222222", got)

	// Remove -> ErrNoRows again.
	s.Require().NoError(s.Q.DeleteCommandLogChannel(ctx))
	_, err = s.Q.GetCommandLogChannel(ctx)
	s.ErrorIs(err, pgx.ErrNoRows)
}

// TestGameConfig covers the game_config key/value store backing the /inv
// create defaults (migration 000029).
func (s *ConfigTestSuite) TestGameConfig() {
	ctx := context.Background()
	const key = "unit_test_config_key"

	s.Require().NoError(s.Q.DeleteGameConfig(ctx, key))
	defer s.Q.DeleteGameConfig(ctx, key)

	// Missing key -> ErrNoRows.
	_, err := s.Q.GetGameConfig(ctx, key)
	s.ErrorIs(err, pgx.ErrNoRows)

	// Upsert + read.
	row, err := s.Q.UpsertGameConfig(ctx, models.UpsertGameConfigParams{Key: key, Value: "200"})
	s.Require().NoError(err)
	s.Equal(key, row.Key)
	s.Equal("200", row.Value)

	got, err := s.Q.GetGameConfig(ctx, key)
	s.Require().NoError(err)
	s.Equal("200", got)

	// Upsert updates in place.
	_, err = s.Q.UpsertGameConfig(ctx, models.UpsertGameConfigParams{Key: key, Value: "250"})
	s.Require().NoError(err)
	got, err = s.Q.GetGameConfig(ctx, key)
	s.Require().NoError(err)
	s.Equal("250", got)

	// Listed among all config entries.
	all, err := s.Q.ListGameConfig(ctx)
	s.Require().NoError(err)
	found := false
	for _, cfg := range all {
		if cfg.Key == key {
			found = true
		}
	}
	s.True(found, "game config key should appear in ListGameConfig")

	// Delete -> ErrNoRows.
	s.Require().NoError(s.Q.DeleteGameConfig(ctx, key))
	_, err = s.Q.GetGameConfig(ctx, key)
	s.True(errors.Is(err, pgx.ErrNoRows))
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
