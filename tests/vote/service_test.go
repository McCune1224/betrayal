package vote

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
	votesvc "github.com/mccune1224/betrayal/internal/services/vote"
	"github.com/mccune1224/betrayal/tests/testutil"
	"github.com/stretchr/testify/suite"
)

const (
	voterID  = int64(100000000000000001)
	targetID = int64(100000000000000002)
	otherID  = int64(100000000000000003)
)

// VoteServiceSuite exercises the vote-casting rules against the LOCAL
// database: player validation, the one-vote-per-voter-per-cycle upsert, and
// per-cycle tallies.
type VoteServiceSuite struct {
	suite.Suite
	DB  *pgxpool.Pool
	Q   *models.Queries
	svc *votesvc.Service
}

func (s *VoteServiceSuite) SetupSuite() {
	s.DB = testutil.NewTestPool(s.T())
	s.Q = models.New(s.DB)
	s.svc = votesvc.New(s.DB)
}

func (s *VoteServiceSuite) SetupTest() {
	testutil.TruncateAll(s.T(), s.DB)
	ctx := context.Background()

	role, err := s.Q.CreateRole(ctx, models.CreateRoleParams{
		Name: "Mafia", Description: "boss", Alignment: models.AlignmentEVIL,
	})
	s.Require().NoError(err)

	for _, id := range []int64{voterID, targetID, otherID} {
		_, err := s.Q.CreatePlayer(ctx, models.CreatePlayerParams{
			ID:        id,
			RoleID:    pgtype.Int4{Int32: role.ID, Valid: true},
			Alive:     true,
			Coins:     200,
			CoinBonus: pgtype.Numeric{},
			Luck:      0,
			Alignment: models.AlignmentEVIL,
		})
		s.Require().NoError(err)
	}
}

func (s *VoteServiceSuite) TestCastVotePersists() {
	ctx := context.Background()

	vote, err := s.svc.CastVote(ctx, voterID, targetID, 1, pgtype.Text{Valid: false})
	s.Require().NoError(err)
	s.Equal(voterID, vote.VoterID)
	s.Equal(targetID, vote.TargetID)
	s.Equal(int32(1), vote.Weight)

	got, err := s.Q.GetVoteByVoterAndCycle(ctx, models.GetVoteByVoterAndCycleParams{
		VoterID: voterID, CycleDay: 0, IsElimination: false,
	})
	s.Require().NoError(err)
	s.Equal(targetID, got.TargetID)
}

func (s *VoteServiceSuite) TestCastVoteUpsertsSameVoterSameCycle() {
	ctx := context.Background()

	_, err := s.svc.CastVote(ctx, voterID, targetID, 1, pgtype.Text{Valid: false})
	s.Require().NoError(err)

	// Same voter, same cycle: the vote is REPLACED, not duplicated.
	_, err = s.svc.CastVote(ctx, voterID, otherID, 1, pgtype.Text{Valid: false})
	s.Require().NoError(err)

	votes, err := s.Q.ListVotesByVoter(ctx, voterID)
	s.Require().NoError(err)
	s.Len(votes, 1)
	s.Equal(otherID, votes[0].TargetID)
}

func (s *VoteServiceSuite) TestCastVoteRejectsNonPlayerVoter() {
	ctx := context.Background()

	_, err := s.svc.CastVote(ctx, 999999999, targetID, 1, pgtype.Text{Valid: false})
	s.Require().Error(err)
	s.True(errors.Is(err, votesvc.ErrNotAPlayer))
}

func (s *VoteServiceSuite) TestCastVoteRejectsNonPlayerTarget() {
	ctx := context.Background()

	_, err := s.svc.CastVote(ctx, voterID, 999999999, 1, pgtype.Text{Valid: false})
	s.Require().Error(err)
	s.True(errors.Is(err, votesvc.ErrNotAPlayer))
}

func (s *VoteServiceSuite) TestIsPlayer() {
	ctx := context.Background()
	s.True(s.svc.IsPlayer(ctx, voterID))
	s.False(s.svc.IsPlayer(ctx, 999999999))
}

func (s *VoteServiceSuite) TestTalliesSumWeightsPerTarget() {
	ctx := context.Background()

	// Two votes for target, one for other, one weighted vote.
	_, err := s.svc.CastVote(ctx, voterID, targetID, 1, pgtype.Text{Valid: false})
	s.Require().NoError(err)
	_, err = s.svc.CastVote(ctx, otherID, targetID, 1, pgtype.Text{Valid: false})
	s.Require().NoError(err)

	tallies, err := s.svc.Tallies(ctx)
	s.Require().NoError(err)
	s.Require().Len(tallies, 1)
	s.Equal(targetID, tallies[0].TargetID)
	s.Equal(int32(2), tallies[0].TotalVotes)
	s.Equal(int64(2), tallies[0].VoteCount)
}

func (s *VoteServiceSuite) TestTalliesOnlyCurrentCycle() {
	ctx := context.Background()

	// Vote in the current cycle (day 0).
	_, err := s.svc.CastVote(ctx, voterID, targetID, 1, pgtype.Text{Valid: false})
	s.Require().NoError(err)

	// Simulate a previous cycle's vote directly.
	_, err = s.Q.UpsertVote(ctx, models.UpsertVoteParams{
		VoterID: otherID, TargetID: targetID,
		CycleDay: 1, IsElimination: true,
		Weight: 5, Context: pgtype.Text{Valid: false},
	})
	s.Require().NoError(err)

	tallies, err := s.svc.Tallies(ctx)
	s.Require().NoError(err)
	s.Require().Len(tallies, 1)
	s.Equal(int32(1), tallies[0].TotalVotes)
}

func TestVoteServiceSuite(t *testing.T) {
	suite.Run(t, new(VoteServiceSuite))
}
