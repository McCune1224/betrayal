package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/zekrotja/ken"
)

type Player struct {
	models data.Models
}

func (p *Player) SetModels(models data.Models) {
	p.models = models
}

var _ ken.SlashCommand = (*Player)(nil)

// Description implements ken.SlashCommand.
func (*Player) Description() string {
	return "Player details"
}

// Name implements ken.SlashCommand.
func (*Player) Name() string {
	return "player"
}

// Options implements ken.SlashCommand.
func (*Player) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add a player with a role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "name",
					Description: "Name of the player",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Name of the role",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "get",
			Description: "Get a player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "name",
					Description: "Name of the player",
					Required:    true,
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (p *Player) Run(ctx ken.Context) (err error) {
	err = ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "add", Run: p.add},
		ken.SubCommandHandler{Name: "get", Run: p.get},
	)
	return err
}

func (p *Player) add(ctx ken.SubCommandContext) (err error) {
	args := ctx.Options()
	name := args.Get(0).UserValue(ctx)
	roleArg := args.Get(1).StringValue()

	role, err := p.models.Roles.GetByName(roleArg)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(err.Error(), "Role not found")
		return err
	}

	player := data.Player{
		DiscordID: name.ID,
		RoleID:    role.ID,
		Coins:     0,
	}

	playerID, err := p.models.Players.Insert(&player)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(err.Error(), "Error adding player")
		return err
	}

	err = ctx.Respond(&discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("All goodie in the hoodie, %s, %d", name.Username, playerID),
		},
	})

	return err
}

func (p *Player) get(ctx ken.SubCommandContext) (err error) {
	args := ctx.Options()
	name := args.Get(0).UserValue(ctx)

	player, err := p.models.Players.GetByDiscordID(name.ID)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(err.Error(), "Player not found")
		return err
	}

	role, err := p.models.Roles.Get(player.RoleID)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(err.Error(), "Player not found")
		return err
	}

	err = ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title: fmt.Sprintf("Player %s", name.Username),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Role",
				Value:  role.Name,
				Inline: true,
			},
			{
				Name:   "Coins",
				Value:  fmt.Sprintf("%d", player.Coins),
				Inline: true,
			},
		},
	})

	return
}

// Version implements ken.SlashCommand.
func (*Player) Version() string {
	return "1.0.0"
}
