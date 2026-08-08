package channels

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

func TestCommandLogChannelRejectsMalformedOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  *ken.CommandOption
	}{
		{name: "missing", opt: nil},
		{name: "wrong type", opt: &ken.CommandOption{ApplicationCommandInteractionDataOption: &discordgo.ApplicationCommandInteractionDataOption{Type: discordgo.ApplicationCommandOptionString, Value: "channel-1"}}},
		{name: "missing value", opt: &ken.CommandOption{ApplicationCommandInteractionDataOption: &discordgo.ApplicationCommandInteractionDataOption{Type: discordgo.ApplicationCommandOptionChannel}}},
		{name: "wrong value", opt: &ken.CommandOption{ApplicationCommandInteractionDataOption: &discordgo.ApplicationCommandInteractionDataOption{Type: discordgo.ApplicationCommandOptionChannel, Value: 42}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := commandLogChannel(tt.opt)
			if err == nil {
				t.Fatal("commandLogChannel error = nil")
			}
			if channel != nil {
				t.Fatalf("channel = %#v, want nil", channel)
			}
		})
	}
}

func TestCommandLogChannelReturnsSafeID(t *testing.T) {
	option := &ken.CommandOption{ApplicationCommandInteractionDataOption: &discordgo.ApplicationCommandInteractionDataOption{
		Type:  discordgo.ApplicationCommandOptionChannel,
		Value: "channel-1",
	}}
	channel, err := commandLogChannel(option)
	if err != nil {
		t.Fatal(err)
	}
	if channel.ID != "channel-1" {
		t.Fatalf("channel ID = %q, want channel-1", channel.ID)
	}
}
