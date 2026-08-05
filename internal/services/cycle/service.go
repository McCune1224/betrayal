// Package cycle implements the game-cycle rules for the Betrayal bot.
// The ken /cycle handlers stay thin; transition logic lives here so it can be
// unit-tested against the local DB without Discord.
package cycle

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
)

// Service is the DB-backed cycle engine used by the /cycle command.
type Service struct {
	pool *pgxpool.Pool
}

// New returns a cycle Service backed by pool.
func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Current returns the persisted game cycle.
func (s *Service) Current(ctx context.Context) (models.GameCycle, error) {
	return models.New(s.pool).GetCycle(ctx)
}

// Set hard-sets the cycle to the given phase and day.
func (s *Service) Set(ctx context.Context, isElimination bool, day int32) (models.GameCycle, error) {
	q := models.New(s.pool)
	curr, err := q.GetCycle(ctx)
	if err != nil {
		return curr, err
	}
	return q.UpdateCycle(ctx, models.UpdateCycleParams{
		ID:            curr.ID,
		Day:           day,
		IsElimination: isElimination,
	})
}

// Increment advances the cycle by one phase and persists it:
// Day 0 -> Day 1; Day n -> Elimination n; Elimination n -> Day n+1.
func (s *Service) Increment(ctx context.Context) (models.GameCycle, error) {
	q := models.New(s.pool)
	curr, err := q.GetCycle(ctx)
	if err != nil {
		return curr, err
	}
	next := NextCycle(curr)
	return q.UpdateCycle(ctx, models.UpdateCycleParams{
		ID:            curr.ID,
		Day:           next.Day,
		IsElimination: next.IsElimination,
	})
}

// NextCycle returns the cycle that follows curr (pure, unit-testable).
func NextCycle(curr models.GameCycle) models.GameCycle {
	if curr.Day == 0 {
		return models.GameCycle{ID: curr.ID, IsElimination: false, Day: 1}
	}
	if curr.IsElimination {
		return models.GameCycle{ID: curr.ID, IsElimination: false, Day: curr.Day + 1}
	}
	return models.GameCycle{ID: curr.ID, IsElimination: true, Day: curr.Day}
}

// FormatMessage returns the broadcast text announcing a cycle transition
// (pure, unit-testable).
func FormatMessage(c models.GameCycle) string {
	// Handle elimination cycle
	if c.IsElimination {
		return fmt.Sprintf("# === END OF DAY %d ===\n# === START OF ELIMINATION %d ===",
			c.Day, c.Day)
	}

	// Handle day 0 transition
	if c.Day-1 == 0 {
		return fmt.Sprintf("# === END OF DAY 0 ===\n# === START OF DAY %d ===",
			c.Day)
	}

	// Handle regular elimination to day transition
	return fmt.Sprintf("# === END OF ELIMINATION %d ===\n# === START OF DAY %d ===",
		c.Day-1, c.Day)
}
