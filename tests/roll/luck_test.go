package roll

import (
	"testing"

	"github.com/mccune1224/betrayal/internal/models"
	rollsvc "github.com/mccune1224/betrayal/internal/services/roll"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRollRarityLevel pins the deterministic luck-table mapping: a fixed
// (level, roll) must always yield the same rarity. These boundaries were
// derived from the chance curves in internal/services/roll/luck.go.
func TestRollRarityLevel(t *testing.T) {
	tests := []struct {
		name  string
		level float64
		roll  float64
		want  models.Rarity
	}{
		// Level 0: common .80 / uncommon .15 / rare .02 / epic .015 / legendary .01 / mythical .005
		{"level 0 roll 0.50", 0, 0.50, models.RarityCOMMON},
		{"level 0 roll 0.79", 0, 0.79, models.RarityCOMMON},
		{"level 0 roll 0.90", 0, 0.90, models.RarityUNCOMMON},
		{"level 0 roll 0.96", 0, 0.96, models.RarityRARE},
		{"level 0 roll 0.98", 0, 0.98, models.RarityEPIC},
		{"level 0 roll 0.99", 0, 0.99, models.RarityLEGENDARY},
		{"level 0 roll 0.999", 0, 0.999, models.RarityMYTHICAL},
		// Level 50: common/uncommon floor to 0; rare .47 / epic .265 / legendary .135 / mythical .13
		{"level 50 roll 0.10", 50, 0.10, models.RarityRARE},
		{"level 50 roll 0.50", 50, 0.50, models.RarityEPIC},
		{"level 50 roll 0.80", 50, 0.80, models.RarityLEGENDARY},
		{"level 50 roll 0.95", 50, 0.95, models.RarityMYTHICAL},
		// Level 100: epic .495 / legendary .26 / mythical .255
		{"level 100 roll 0.30", 100, 0.30, models.RarityEPIC},
		{"level 100 roll 0.60", 100, 0.60, models.RarityLEGENDARY},
		{"level 100 roll 0.90", 100, 0.90, models.RarityMYTHICAL},
		// Anything above level 397 is always mythical.
		{"level 500 roll 0.00", 500, 0.0, models.RarityMYTHICAL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rollsvc.RollRarityLevel(tt.level, tt.roll)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBaseLuckChances pins the level-0 base rates.
func TestBaseLuckChances(t *testing.T) {
	assert.InDelta(t, 0.80, rollsvc.CommonLuckChance(0), 0.0001)
	assert.InDelta(t, 0.15, rollsvc.UncommonLuckChance(0), 0.0001)
	assert.InDelta(t, 0.02, rollsvc.RareLuckChance(0), 0.0001)
	assert.InDelta(t, 0.015, rollsvc.EpicLuckChance(0), 0.0001)
	assert.InDelta(t, 0.01, rollsvc.LegendaryLuckChance(0), 0.0001)
	assert.InDelta(t, 0.005, rollsvc.MythicalLuckChance(0), 0.0001)
}

// TestLuckChanceEdgeCases pins documented quirks of the curves.
func TestLuckChanceEdgeCases(t *testing.T) {
	// Rare is constant 0.49 at level 48 (documented edge case).
	assert.InDelta(t, 0.49, rollsvc.RareLuckChance(48), 0.0001)

	// Level 50 curves sum to ~1.0 (common/uncommon floored to 0).
	sum := rollsvc.RareLuckChance(50) + rollsvc.EpicLuckChance(50) +
		rollsvc.LegendaryLuckChance(50) + rollsvc.MythicalLuckChance(50)
	assert.InDelta(t, 1.0, sum, 0.0001)

	// Mythical is capped at 1.
	assert.InDelta(t, 1.0, rollsvc.MythicalLuckChance(1000), 0.0001)
}

// TestRollAtRarity verifies that roll-at-minimum-rarity only ever returns an
// allowed rarity (rand-driven, so membership over many draws).
func TestRollAtRarity(t *testing.T) {
	allowed := []models.Rarity{models.RarityRARE, models.RarityEPIC, models.RarityLEGENDARY, models.RarityMYTHICAL}
	for i := 0; i < 100; i++ {
		got := rollsvc.RollAtRarity(0, allowed)
		require.Contains(t, allowed, got)
	}
}
