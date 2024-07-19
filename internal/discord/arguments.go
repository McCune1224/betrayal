package discord

import "github.com/bwmarrin/discordgo"

// Common Arugment types used within the application.
// Thse are really just here because indentation hell is a thing with discordgo.
// I'm not sure if I like this or not, but it's better than the alternative.

// Require an option called "user" within the command.
func UserCommandArg(required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionUser,
		Name:        "user",
		Description: "User to target",
		Required:    required,
	}
}

func StatusCommandArg(name, description string, required bool) *discordgo.ApplicationCommandOption {
	// WARNING: HAHAHHAHAHAHAHAHAHA IM SO SORRY
	statuses := []string{
		"Cursed",
		"Death Cursed",
		"Frozen",
		"Paralyzed",
		"Burned",
		"Empowered",
		"Drunk",
		"Restrained",
		"Disabled",
		"Blackmailed",
		"Despaired",
		"Madness",
		"Lucky",
		"Unlucky",
	}
	choices := []*discordgo.ApplicationCommandOptionChoice{}
	for _, status := range statuses {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{Name: status, Value: status})
	}

	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: description,
		Required:    required,
		Choices:     choices,
	}
}

func StringCommandArg(
	name string,
	description string,
	required bool,
) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

func IntCommandArg(
	name string,
	description string,
	required bool,
) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

func BoolCommandArg(
	name string,
	description string,
	required bool,
) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionBoolean,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// Require an option called "channel" within the command.
func ChannelCommandArg(required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionChannel,
		Name:        "channel",
		Description: "Channel to target",
		Required:    required,
	}
}
