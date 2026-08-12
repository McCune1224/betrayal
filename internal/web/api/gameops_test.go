package api

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestChannelEntryUsesReadableDiscordLabel(t *testing.T) {
	session := &discordgo.Session{State: &discordgo.State{Ready: discordgo.Ready{Guilds: []*discordgo.Guild{{Channels: []*discordgo.Channel{{ID: "123456789012345678", Name: "vote-funnel"}}}}}}}
	handler := NewChannelsHandler(nil, session)

	entry := handler.channelEntry("Vote Channel", "vote", "123456789012345678")

	if entry.Label != "#vote-funnel" {
		t.Fatalf("label = %q, want %q", entry.Label, "#vote-funnel")
	}
	if entry.Label == entry.ChannelID {
		t.Fatal("human-facing label must not be the raw Discord channel ID")
	}
}

func TestChannelEntryUsesReadableLocalFallback(t *testing.T) {
	handler := NewChannelsHandler(nil, nil)

	entry := handler.channelEntry("Vote Channel", "vote", "123456789012345678")

	if entry.Label != "Configured Discord channel" {
		t.Fatalf("label = %q, want readable local fallback", entry.Label)
	}
	if entry.Label == entry.ChannelID {
		t.Fatal("local fallback must not expose the raw Discord channel ID")
	}
}
