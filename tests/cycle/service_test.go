package cycle

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	cyclesvc "github.com/mccune1224/betrayal/internal/services/cycle"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestNextCycle pins the pure transition rules:
// Day 0 -> Day 1; Day n -> Elimination n; Elimination n -> Day n+1.
func TestNextCycle(t *testing.T) {
	tests := []struct {
		name string
		curr models.GameCycle
		want models.GameCycle
	}{
		{"day 0 -> day 1", models.GameCycle{ID: 1, IsElimination: false, Day: 0}, models.GameCycle{ID: 1, IsElimination: false, Day: 1}},
		{"day 2 -> elimination 2", models.GameCycle{ID: 1, IsElimination: false, Day: 2}, models.GameCycle{ID: 1, IsElimination: true, Day: 2}},
		{"elimination 3 -> day 4", models.GameCycle{ID: 1, IsElimination: true, Day: 3}, models.GameCycle{ID: 1, IsElimination: false, Day: 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cyclesvc.NextCycle(tt.curr))
		})
	}
}

// TestFormatMessage pins the broadcast text for each transition type.
func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name  string
		cycle models.GameCycle
		want  string
	}{
		{
			"day 1 start",
			models.GameCycle{IsElimination: false, Day: 1},
			"# === END OF DAY 0 ===\n# === START OF DAY 1 ===",
		},
		{
			"regular day start",
			models.GameCycle{IsElimination: false, Day: 4},
			"# === END OF ELIMINATION 3 ===\n# === START OF DAY 4 ===",
		},
		{
			"elimination start",
			models.GameCycle{IsElimination: true, Day: 4},
			"# === END OF DAY 4 ===\n# === START OF ELIMINATION 4 ===",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cyclesvc.FormatMessage(tt.cycle))
		})
	}
}

// CycleServiceSuite exercises the DB-backed cycle transitions against the
// LOCAL database. TruncateAll re-seeds the game_cycle Day-0 row, so every
// test starts from the migration's initial state.
type CycleServiceSuite struct {
	suite.Suite
	DB  *pgxpool.Pool
	svc *cyclesvc.Service
}

func (s *CycleServiceSuite) SetupSuite() {
	s.DB = testutil.NewTestPool(s.T())
	s.svc = cyclesvc.New(s.DB)
}

func (s *CycleServiceSuite) SetupTest() {
	testutil.TruncateAll(s.T(), s.DB)
}

func (s *CycleServiceSuite) TestCurrentStartsAtDayZero() {
	cycle, err := s.svc.Current(context.Background())
	s.Require().NoError(err)
	s.Equal(int32(0), cycle.Day)
	s.False(cycle.IsElimination)
}

func (s *CycleServiceSuite) TestIncrementDayZeroToDayOne() {
	next, err := s.svc.Increment(context.Background())
	s.Require().NoError(err)
	s.Equal(int32(1), next.Day)
	s.False(next.IsElimination)

	// Persisted.
	cycle, err := s.svc.Current(context.Background())
	s.Require().NoError(err)
	s.Equal(next, cycle)
}

func (s *CycleServiceSuite) TestIncrementDayToElimination() {
	_, err := s.svc.Set(context.Background(), false, 3)
	s.Require().NoError(err)

	next, err := s.svc.Increment(context.Background())
	s.Require().NoError(err)
	s.Equal(int32(3), next.Day)
	s.True(next.IsElimination)
}

func (s *CycleServiceSuite) TestIncrementEliminationToNextDay() {
	_, err := s.svc.Set(context.Background(), true, 3)
	s.Require().NoError(err)

	next, err := s.svc.Increment(context.Background())
	s.Require().NoError(err)
	s.Equal(int32(4), next.Day)
	s.False(next.IsElimination)
}

func (s *CycleServiceSuite) TestSetHardSetsPhaseAndDay() {
	cycle, err := s.svc.Set(context.Background(), true, 7)
	s.Require().NoError(err)
	s.Equal(int32(7), cycle.Day)
	s.True(cycle.IsElimination)
}

func (s *CycleServiceSuite) TestFormatMessageForPersistedCycle() {
	_, err := s.svc.Set(context.Background(), false, 2)
	s.Require().NoError(err)

	cycle, err := s.svc.Current(context.Background())
	s.Require().NoError(err)

	msg := cyclesvc.FormatMessage(cycle)
	require.True(s.T(), strings.Contains(msg, "DAY 2"))
	require.True(s.T(), strings.Contains(msg, "ELIMINATION"))
}

func TestCycleServiceSuite(t *testing.T) {
	suite.Run(t, new(CycleServiceSuite))
}
