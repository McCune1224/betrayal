package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestMessageURLUsesReferenceGuild(t *testing.T) {
	ref := &discordgo.MessageReference{
		GuildID:   "guild-2",
		ChannelID: "channel-2",
		MessageID: "message-2",
	}

	got := MessageURL(ref)
	want := "https://discord.com/channels/guild-2/channel-2/message-2"
	if got != want {
		t.Fatalf("message URL = %q, want %q", got, want)
	}
}

func TestMessageURLUsesDirectMessageMarkerWithoutGuild(t *testing.T) {
	ref := &discordgo.MessageReference{ChannelID: "channel-1", MessageID: "message-1"}
	if got, want := MessageURL(ref), "https://discord.com/channels/@me/channel-1/message-1"; got != want {
		t.Fatalf("message URL = %q, want %q", got, want)
	}
}
