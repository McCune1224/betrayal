package whisper

import (
	"testing"

	whispersvc "github.com/mccune1224/betrayal/internal/services/whisper"
)

// TestSecureRollerHitBoundaryIsExactlyFivePercent pins the float boundary that
// makes Hit(SuspicionChance) exactly 500/10000 = 5.00%: Intn(10000) < 500.
func TestSecureRollerHitBoundaryIsExactlyFivePercent(t *testing.T) {
	bound := int(whispersvc.SuspicionChance * 10000)
	if bound != 500 {
		t.Fatalf("Hit boundary = %d, want 500 so Intn(10000) < bound is exactly 5%%", bound)
	}
}

// TestSecureRollerHitsSuspicionChanceAtFivePercent samples the production
// crypto roller against the real SuspicionChance constant. At 200k samples the
// standard error is ~0.05%, so the 4%-6% band is >20 sigma: it cannot flake,
// yet fails hard if the chance is ever misconfigured (0%, 10%, always-true).
func TestSecureRollerHitsSuspicionChanceAtFivePercent(t *testing.T) {
	const samples = 200_000
	roller := secureRoller{}
	hits := 0
	for i := 0; i < samples; i++ {
		if roller.Hit(whispersvc.SuspicionChance) {
			hits++
		}
	}
	rate := float64(hits) / samples
	if rate < 0.04 || rate > 0.06 {
		t.Fatalf("secureRoller.Hit(SuspicionChance) fired %d/%d = %.4f, want ~0.05", hits, samples, rate)
	}
}

// TestSecureRollerIntnIsUniform verifies the doubt-message picker draws from
// the whole pool uniformly. At 200k draws per bucket the fraction band
// 14%-19% around 1/6 is >15 sigma, so it cannot flake while still rejecting a
// picker stuck on a single index.
func TestSecureRollerIntnIsUniform(t *testing.T) {
	const n = 6
	const samples = 200_000
	roller := secureRoller{}
	counts := make([]int, n)
	for i := 0; i < samples; i++ {
		counts[roller.Intn(n)]++
	}
	for i, c := range counts {
		frac := float64(c) / samples
		if frac < 0.14 || frac > 0.19 {
			t.Fatalf("secureRoller.Intn(%d) bucket %d = %d/%d = %.4f, want ~1/6", n, i, c, samples, frac)
		}
	}
}

func TestSecureRollerIntnDegeneratesToZeroForSingleEntry(t *testing.T) {
	roller := secureRoller{}
	for i := 0; i < 100; i++ {
		if got := roller.Intn(1); got != 0 {
			t.Fatalf("Intn(1) = %d, want 0 (single-entry pool must index the only message)", got)
		}
	}
}
