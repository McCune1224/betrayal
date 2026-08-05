package roll

import (
	"context"
	"math/rand"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
)

// Service is the DB-backed roll engine used by the /roll command. The ken
// handlers stay thin; all draw decision logic lives here so it can be tested
// against the local DB without Discord.
type Service struct {
	pool *pgxpool.Pool
}

// New returns a roll Service backed by pool.
func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// RollLevel draws a rarity for a luck level using a fresh random roll.
func RollLevel(level float64) models.Rarity {
	return RollRarityLevel(level, rand.Float64())
}

// RollItemByRarity returns a random item of exactly rarity r.
func (s *Service) RollItemByRarity(ctx context.Context, r models.Rarity) (models.Item, error) {
	return models.New(s.pool).GetRandomItemByRarity(ctx, r)
}

// RollItemAtMinimumRarity returns a random item of rarity >= min (never UNIQUE).
func (s *Service) RollItemAtMinimumRarity(ctx context.Context, min models.Rarity) (models.Item, error) {
	return models.New(s.pool).GetRandomItemByMinimumRarity(ctx, min)
}

// RollAnyAbilityByRarity returns a random any-ability of exactly rarity r.
func (s *Service) RollAnyAbilityByRarity(ctx context.Context, r models.Rarity) (models.AbilityInfo, error) {
	return models.New(s.pool).GetRandomAnyAbilityByRarity(ctx, r)
}

// RollAnyAbilityAtMinimumRarity returns a random any-ability of rarity >= min.
func (s *Service) RollAnyAbilityAtMinimumRarity(ctx context.Context, min models.Rarity) (models.AbilityInfo, error) {
	return models.New(s.pool).GetRandomAnyAbilityByMinimumRarity(ctx, min)
}

// RollAnyAbilityIncludingRoleSpecific returns a random any-ability of rarity r,
// including abilities tied to the player's role.
func (s *Service) RollAnyAbilityIncludingRoleSpecific(ctx context.Context, r models.Rarity, roleID int32) (models.AbilityInfo, error) {
	return models.New(s.pool).GetRandomAnyAbilityIncludingRoleSpecific(ctx, models.GetRandomAnyAbilityIncludingRoleSpecificParams{
		Rarity: r,
		RoleID: roleID,
	})
}
