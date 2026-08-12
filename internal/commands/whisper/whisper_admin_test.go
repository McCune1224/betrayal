package whisper

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestOptionsExposeAdminWhisperManagement(t *testing.T) {
	options := (&Whisper{}).Options()
	message := findNestedOption(options, "message")
	if message == nil || message.Type != discordgo.ApplicationCommandOptionString {
		t.Fatal("/whisper must expose a message option")
	}
	if findOption(options, "admin", discordgo.ApplicationCommandOptionSubCommandGroup) != nil {
		t.Fatal("player /whisper command must not expose admin subcommands")
	}
	admin := (&WhisperAdmin{}).Options()
	for _, name := range []string{"group-list", "group-create", "group-delete", "member-add", "member-remove", "message-list", "message-create", "message-update", "message-enable", "message-disable", "message-delete"} {
		if findOption(admin, name, discordgo.ApplicationCommandOptionSubCommand) == nil {
			t.Errorf("missing /whisper-admin %s subcommand", name)
		}
	}
	for _, name := range []string{"member-add", "member-remove"} {
		option := findOption(admin, name, discordgo.ApplicationCommandOptionSubCommand)
		player := findNestedOption(option.Options, "user")
		if player == nil || player.Type != discordgo.ApplicationCommandOptionUser {
			t.Errorf("/whisper-admin %s must target a Discord user, not a raw ID", name)
		}
	}
}

func findOption(options []*discordgo.ApplicationCommandOption, name string, kind discordgo.ApplicationCommandOptionType) *discordgo.ApplicationCommandOption {
	for _, option := range options {
		if option != nil && option.Name == name && option.Type == kind {
			return option
		}
	}
	return nil
}

func findNestedOption(options []*discordgo.ApplicationCommandOption, name string) *discordgo.ApplicationCommandOption {
	for _, option := range options {
		if option != nil && option.Name == name {
			return option
		}
	}
	return nil
}
