package roll

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
)

// Base luck types chances (at level 0)
var (
	// 80%
	commonLuck = 0.800
	// 15%
	uncommonLuck = 0.150
	// 2%
	rareLuck = 0.020
	// 1.5%
	epicLuck = 0.015
	// 1%
	legendaryLuck = 0.010
	// 0.5%
	mythicalLuck = 0.005

	// list of luck types in order of rarity
	rarityPriorities = []models.Rarity{models.RarityCOMMON, models.RarityUNCOMMON, models.RarityRARE, models.RarityEPIC, models.RarityLEGENDARY, models.RarityMYTHICAL}
)

func commonLuckChance(level float64) float64 {
	scale := -0.050 * float64(level)
	chance := commonLuck + scale

	// round to 4th decimal place
	chance = math.Round(chance*10000) / 10000
	return math.Max(chance, 0)
}

// helper to round down to 4th decimal place and return max of 0
func sanatized(num float64) float64 {
	r := math.Round(num*10000) / 10000
	return math.Max(r, 0)
}

func uncommonLuckChance(level float64) float64 {
	flipLevel := 16.00
	neg := 0.02
	pos := 0.03
	if level > flipLevel {
		return sanatized(uncommonLuckChance(flipLevel) - (level-flipLevel)*neg)
	}
	scale := pos * float64(level)
	chance := uncommonLuck + scale
	return sanatized(chance)
}

func rareLuckChance(level float64) float64 {
	// rare has random edge case at luck level 48 where it is constant at .49
	if level == 48 {
		return 0.49
	}
	flipLevel := 48.00
	neg := 0.01
	pos := 0.01
	if level > flipLevel {
		return sanatized(rareLuckChance(flipLevel) - (level-flipLevel)*neg)
	}
	scale := pos * float64(level)
	chance := rareLuck + scale
	return sanatized(chance)
}

func epicLuckChance(level float64) float64 {
	flipLevel := 98.00
	neg := 0.005
	pos := 0.005
	if level > flipLevel {
		return sanatized(epicLuckChance(flipLevel) - (level-flipLevel)*neg)
	}
	scale := pos * float64(level)
	chance := epicLuck + scale
	return sanatized(chance)
}

func legendaryLuckChance(level float64) float64 {
	flipLevel := 198.00
	neg := 0.0025
	pos := 0.0025
	if level > flipLevel {
		return sanatized(legendaryLuckChance(flipLevel) - (level-flipLevel)*neg)
	}
	scale := pos * float64(level)
	chance := legendaryLuck + scale
	return sanatized(chance)
}

func mythicalLuckChance(level float64) float64 {
	// mythical just scales 0.25% per level
	scale := 0.0025 * float64(level)
	if mythicalLuck+scale > 1 {
		return 1
	}
	return sanatized(mythicalLuck + scale)
}

// pick a random number between 0 and 1 and select the luck type based on the number
func RollRarityLevel(level float64, roll float64) models.Rarity {
	if level > 397 {
		return models.RarityMYTHICAL
	}

	cc := commonLuckChance(level)
	uc := uncommonLuckChance(level)
	rc := rareLuckChance(level)
	ec := epicLuckChance(level)
	lc := legendaryLuckChance(level)
	mc := mythicalLuckChance(level)

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

func rollAtRarity(level float64, allowedRarities []models.Rarity) models.Rarity {
	roll := rand.Float64()
	// roll until we get a rarity that is allowed
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

// display chances of each type at a given level
func tableView(level float64, roll float64) string {
	c := commonLuckChance(level)
	uc := uncommonLuckChance(level)
	rc := rareLuckChance(level)
	ec := epicLuckChance(level)
	lc := legendaryLuckChance(level)
	mc := mythicalLuckChance(level)

	outcome := RollRarityLevel(level, roll)
	// print as human readable percentage
	s := fmt.Sprintf(
		"LEVEL - %v\n%.2f%% common\n%.2f%% uncommon\n%.2f%% rare\n%.2f%% epic\n%.2f%% legendary\n%.2f%% mythical\n\nroll: %v - %v",
		level,
		c*100,
		uc*100,
		rc*100,
		ec*100,
		lc*100,
		mc*100,
		outcome,
		roll,
	)

	return discord.Code(s)
}
