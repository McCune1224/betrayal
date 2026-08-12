// Package roll implements the luck/rarity roll engine for the Betrayal bot.
// The pure probability math lives here (unit-testable without Discord or a
// DB); the DB-backed draw methods live in service.go.
package roll

import (
	"math/rand"

	"github.com/mccune1224/betrayal/internal/models"
)

// Base luck type chances (at level 0). The checkpoint table below moves these
// probabilities toward rarer tiers while keeping the total exactly 100%.
var (
	// RarityPriorities lists rarities in order of scarcity.
	RarityPriorities = []models.Rarity{models.RarityCOMMON, models.RarityUNCOMMON, models.RarityRARE, models.RarityEPIC, models.RarityLEGENDARY, models.RarityMYTHICAL}
	luckCheckpoints  = [...]float64{0, 25, 50, 75, 100}
	luckChances      = [...][6]float64{
		{.80, .15, .02, .015, .01, .005},
		{.65, .17, .08, .05, .035, .015},
		{.50, .20, .12, .08, .06, .04},
		{.35, .20, .15, .10, .12, .08},
		{.25, .20, .15, .10, .20, .10},
	}
)

// luckChance linearly interpolates between the approved luck checkpoints.
// Luck is clamped to [0, 100], so it cannot create discontinuities or
// probabilities whose sum exceeds 100%.
func luckChance(level float64, rarity int) float64 {
	if level <= luckCheckpoints[0] {
		return luckChances[0][rarity]
	}
	if level >= luckCheckpoints[len(luckCheckpoints)-1] {
		return luckChances[len(luckChances)-1][rarity]
	}

	for i := 1; i < len(luckCheckpoints); i++ {
		if level <= luckCheckpoints[i] {
			fraction := (level - luckCheckpoints[i-1]) / (luckCheckpoints[i] - luckCheckpoints[i-1])
			return luckChances[i-1][rarity] + fraction*(luckChances[i][rarity]-luckChances[i-1][rarity])
		}
	}
	return luckChances[len(luckChances)-1][rarity]
}

func CommonLuckChance(level float64) float64    { return luckChance(level, 0) }
func UncommonLuckChance(level float64) float64  { return luckChance(level, 1) }
func RareLuckChance(level float64) float64      { return luckChance(level, 2) }
func EpicLuckChance(level float64) float64      { return luckChance(level, 3) }
func LegendaryLuckChance(level float64) float64 { return luckChance(level, 4) }
func MythicalLuckChance(level float64) float64  { return luckChance(level, 5) }

// RollRarityLevel picks a rarity for a luck level given a roll in [0, 1).
// Deterministic for a fixed (level, roll) — the command layer supplies the
// random draw, which keeps this unit-testable.
func RollRarityLevel(level float64, roll float64) models.Rarity {
	cc := CommonLuckChance(level)
	uc := UncommonLuckChance(level)
	rc := RareLuckChance(level)
	ec := EpicLuckChance(level)
	lc := LegendaryLuckChance(level)
	mc := MythicalLuckChance(level)

	if roll < cc {
		return models.RarityCOMMON
	}
	if roll < cc+uc {
		return models.RarityUNCOMMON
	}
	if roll < cc+uc+rc {
		return models.RarityRARE
	}
	if roll < cc+uc+rc+ec {
		return models.RarityEPIC
	}
	if roll < cc+uc+rc+ec+lc {
		return models.RarityLEGENDARY
	}
	if roll < cc+uc+rc+ec+lc+mc {
		return models.RarityMYTHICAL
	}

	// Anything above like 397 is just mythical so just return that
	return models.RarityMYTHICAL
}

// RollAtRarity draws (using math/rand) until a rarity within allowedRarities
// is rolled.
func RollAtRarity(level float64, allowedRarities []models.Rarity) models.Rarity {
	roll := rand.Float64()
	for {
		rarity := RollRarityLevel(level, roll)
		for _, allowed := range allowedRarities {
			if allowed == rarity {
				return rarity
			}
		}
		roll = rand.Float64()
	}
}
