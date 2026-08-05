package main

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestGatewayIntents(t *testing.T) {
	intents := gatewayIntents()

	// Regression: the intents field used to be assigned the
	// PermissionAdministrator permission constant (1<<3 == 8), which happens to
	// equal the emoji intent bit — a permission, not a real intent set.
	if int(intents) == int(discordgo.PermissionAdministrator) {
		t.Fatal("gateway intents must not be the PermissionAdministrator permission constant")
	}

	// Intents needed to keep the Discord state cache populated and to observe
	// message/reaction activity in the bot's guilds.
	required := []discordgo.Intent{
		discordgo.IntentsGuilds,
		discordgo.IntentsGuildMessages,
		discordgo.IntentsGuildMessageReactions,
		discordgo.IntentsGuildEmojis,
		discordgo.IntentsGuildVoiceStates,
	}
	for _, want := range required {
		if intents&want != want {
			t.Errorf("gateway intents missing bit %d", want)
		}
	}

	// Privileged intents (GuildMembers, GuildPresences, MessageContent) require
	// explicit enablement in the Discord developer portal; a bot that connects
	// without them enabled gets disconnected with a 4014. IntentsAllWithoutPrivileged
	// deliberately excludes them.
	privileged := discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences | discordgo.IntentsMessageContent
	if intents&privileged != 0 {
		t.Errorf("gateway intents unexpectedly include privileged bits %d", intents&privileged)
	}
}
