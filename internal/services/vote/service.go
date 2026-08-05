// Package vote implements the vote-casting rules for the Betrayal bot.
// The ken /vote handlers stay thin (session lookups only); the validation and
// upsert decision logic lives here so it can be unit-tested against the local
// DB without Discord.
package vote

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
)

// ErrNotAPlayer is returned when a voter or vote target is not a registered
// player.
var ErrNotAPlayer = errors.New("not a registered player")

// Service is the DB-backed vote engine used by the /vote command.
type Service struct {
	pool *pgxpool.Pool
}

// New returns a vote Service backed by pool.
func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// IsPlayer reports whether id is a registered player.
func (s *Service) IsPlayer(ctx context.Context, id int64) bool {
	_, err := models.New(s.pool).GetPlayer(ctx, id)
	return err == nil
}

// CastVote records (or replaces, via upsert) one player's vote for a target in
// the current game cycle. It validates that both voter and target are
// registered players. Exactly one vote per voter per cycle (day + phase) is
// kept — a second vote replaces the target.
func (s *Service) CastVote(ctx context.Context, voterID, targetID int64, weight int32, contextText pgtype.Text) (models.Vote, error) {
	q := models.New(s.pool)

	if _, err := q.GetPlayer(ctx, voterID); err != nil {
		return models.Vote{}, fmt.Errorf("%w: voter %d", ErrNotAPlayer, voterID)
	}
	if _, err := q.GetPlayer(ctx, targetID); err != nil {
		return models.Vote{}, fmt.Errorf("%w: target %d", ErrNotAPlayer, targetID)
	}

	cycle, err := q.GetCycle(ctx)
	if err != nil {
		return models.Vote{}, fmt.Errorf("get game cycle: %w", err)
	}

	return q.UpsertVote(ctx, models.UpsertVoteParams{
		VoterID:       voterID,
		TargetID:      targetID,
		CycleDay:      cycle.Day,
		IsElimination: cycle.IsElimination,
		Weight:        weight,
		Context:       contextText,
	})
}

// Tallies returns per-target vote totals for the current cycle, highest first.
func (s *Service) Tallies(ctx context.Context) ([]models.GetVoteTalliesByCycleRow, error) {
	q := models.New(s.pool)
	cycle, err := q.GetCycle(ctx)
	if err != nil {
		return nil, err
	}
	return q.GetVoteTalliesByCycle(ctx, models.GetVoteTalliesByCycleParams{
		CycleDay:      cycle.Day,
		IsElimination: cycle.IsElimination,
	})
}
