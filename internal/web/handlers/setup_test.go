package handlers

import "testing"

func TestConfiguredGuildIDPrefersEnvironment(t *testing.T) {
	got := configuredGuildID(func(key string) string {
		if key == "DISCORD_GUILD_ID" {
			return "configured-guild"
		}
		return ""
	})
	if got != "configured-guild" {
		t.Fatalf("configured guild ID = %q, want configured-guild", got)
	}
}

func TestConfiguredGuildIDFallsBackForLegacyWebSetup(t *testing.T) {
	got := configuredGuildID(func(string) string { return "" })
	if got != "1096058997477490861" {
		t.Fatalf("fallback guild ID = %q, want legacy guild", got)
	}
}
