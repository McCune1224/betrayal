package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestInteractionGuildIDUsesCurrentInteractionGuild(t *testing.T) {
	interaction := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{GuildID: "development-guild"}}

	if got := InteractionGuildID(interaction); got != "development-guild" {
		t.Fatalf("interaction guild ID = %q, want development-guild", got)
	}
}

func TestInteractionGuildIDHandlesMissingInteraction(t *testing.T) {
	if got := InteractionGuildID(nil); got != "" {
		t.Fatalf("missing interaction guild ID = %q, want empty", got)
	}
}
