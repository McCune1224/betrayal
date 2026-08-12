package roll

import (
	"testing"

	"github.com/mccune1224/betrayal/internal/models"
	rollsvc "github.com/mccune1224/betrayal/internal/services/roll"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRollRarityLevel pins the deterministic luck-table mapping: a fixed
// (level, roll) must always yield the same rarity.
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
		// Level 50: common .50 / uncommon .20 / rare .12 / epic .08 / legendary .06 / mythical .04
		{"level 50 roll 0.10", 50, 0.10, models.RarityCOMMON},
		{"level 50 roll 0.72", 50, 0.72, models.RarityRARE},
		{"level 50 roll 0.85", 50, 0.85, models.RarityEPIC},
		{"level 50 roll 0.95", 50, 0.95, models.RarityLEGENDARY},
		// Level 100: common .25 / uncommon .20 / rare .15 / epic .10 / legendary .20 / mythical .10
		{"level 100 roll 0.50", 100, 0.50, models.RarityRARE},
		{"level 100 roll 0.85", 100, 0.85, models.RarityLEGENDARY},
		{"level 100 roll 0.95", 100, 0.95, models.RarityMYTHICAL},
		// Luck is capped at the level-100 distribution.
		{"level 500 roll 0.00", 500, 0.0, models.RarityCOMMON},
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

// TestLuckChanceEdgeCases pins the checkpoint distributions and invariants.
func TestLuckChanceEdgeCases(t *testing.T) {
	for _, tt := range []struct {
		level float64
		want  []float64
	}{
		{0, []float64{.80, .15, .02, .015, .01, .005}},
		{25, []float64{.65, .17, .08, .05, .035, .015}},
		{50, []float64{.50, .20, .12, .08, .06, .04}},
		{75, []float64{.35, .20, .15, .10, .12, .08}},
		{100, []float64{.25, .20, .15, .10, .20, .10}},
		{500, []float64{.25, .20, .15, .10, .20, .10}},
	} {
		got := []float64{rollsvc.CommonLuckChance(tt.level), rollsvc.UncommonLuckChance(tt.level), rollsvc.RareLuckChance(tt.level), rollsvc.EpicLuckChance(tt.level), rollsvc.LegendaryLuckChance(tt.level), rollsvc.MythicalLuckChance(tt.level)}
		sum := 0.0
		for i := range got {
			assert.InDelta(t, tt.want[i], got[i], .0001)
			sum += got[i]
		}
		assert.InDelta(t, 1.0, sum, .0001)
	}

	// Intermediate levels are linearly interpolated between checkpoints.
	assert.InDelta(t, .725, rollsvc.CommonLuckChance(12.5), .0001)
	assert.InDelta(t, .185, rollsvc.UncommonLuckChance(37.5), .0001)
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
