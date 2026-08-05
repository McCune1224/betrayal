// Package roll implements the luck/rarity roll engine for the Betrayal bot.
// The pure probability math lives here (unit-testable without Discord or a
// DB); the DB-backed draw methods live in service.go.
package roll

import (
	"math"
	"math/rand"

	"github.com/mccune1224/betrayal/internal/models"
)

// Base luck type chances (at level 0).
var (
	// 80%
	CommonLuck = 0.800
	// 15%
	UncommonLuck = 0.150
	// 2%
	RareLuck = 0.020
	// 1.5%
	EpicLuck = 0.015
	// 1%
	LegendaryLuck = 0.010
	// 0.5%
	MythicalLuck = 0.005

	// RarityPriorities lists rarities in order of scarcity.
	RarityPriorities = []models.Rarity{models.RarityCOMMON, models.RarityUNCOMMON, models.RarityRARE, models.RarityEPIC, models.RarityLEGENDARY, models.RarityMYTHICAL}
)

func CommonLuckChance(level float64) float64 {
	scale := -0.050 * float64(level)
	chance := CommonLuck + scale

	// round to 4th decimal place
	chance = math.Round(chance*10000) / 10000
	return math.Max(chance, 0)
}

// sanatized rounds down to the 4th decimal place and floors at 0.
func sanatized(num float64) float64 {
	r := math.Round(num*10000) / 10000
	return math.Max(r, 0)
}

func UncommonLuckChance(level float64) float64 {
	flipLevel := 16.00
	neg := 0.02
	pos := 0.03
	if level > flipLevel {
		return sanatized(UncommonLuckChance(flipLevel) - (level-flipLevel)*neg)
	}
	scale := pos * float64(level)
	chance := UncommonLuck + scale
	return sanatized(chance)
}

func RareLuckChance(level float64) float64 {
	// rare has random edge case at luck level 48 where it is constant at .49
	if level == 48 {
		return 0.49
	}
	flipLevel := 48.00
	neg := 0.01
	pos := 0.01
	if level > flipLevel {
		return sanatized(RareLuckChance(flipLevel) - (level-flipLevel)*neg)
	}
	scale := pos * float64(level)
	chance := RareLuck + scale
	return sanatized(chance)
}

func EpicLuckChance(level float64) float64 {
	flipLevel := 98.00
	neg := 0.005
	pos := 0.005
	if level > flipLevel {
		return sanatized(EpicLuckChance(flipLevel) - (level-flipLevel)*neg)
	}
	scale := pos * float64(level)
	chance := EpicLuck + scale
	return sanatized(chance)
}

func LegendaryLuckChance(level float64) float64 {
	flipLevel := 198.00
	neg := 0.0025
	pos := 0.0025
	if level > flipLevel {
		return sanatized(LegendaryLuckChance(flipLevel) - (level-flipLevel)*neg)
	}
	scale := pos * float64(level)
	chance := LegendaryLuck + scale
	return sanatized(chance)
}

func MythicalLuckChance(level float64) float64 {
	// mythical just scales 0.25% per level
	scale := 0.0025 * float64(level)
	if MythicalLuck+scale > 1 {
		return 1
	}
	return sanatized(MythicalLuck + scale)
}

// RollRarityLevel picks a rarity for a luck level given a roll in [0, 1).
// Deterministic for a fixed (level, roll) — the command layer supplies the
// random draw, which keeps this unit-testable.
func RollRarityLevel(level float64, roll float64) models.Rarity {
	if level > 397 {
		return models.RarityMYTHICAL
	}

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
