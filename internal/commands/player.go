package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
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
	return discord.DebugCmd + "player"
}

// Options implements ken.SlashCommand.
func (*Player) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add",
			Description: "Add a player with a role",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
				discord.StringCommandArg("role", "Role to assign to the player", true),
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
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	ctx.SetEphemeral(true)
	args := ctx.Options()
	name := args.GetByName("name").UserValue(ctx)
	roleArg := args.GetByName("role").StringValue()

	role, err := p.models.Roles.GetByFuzzy(roleArg)
	if err != nil {
		discord.ErrorMessage(ctx, "Unable to find Role", "No known role of name "+roleArg)
		return err
	}

	player := data.Player{
		DiscordID: name.ID,
		RoleID:    role.ID,
		Coins:     0,
	}

	playerID, err := p.models.Players.Insert(&player)
	if err != nil {
		discord.ErrorMessage(ctx, "Unable to add Player", "Unable to add Player "+name.Username)
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
  if err := ctx.Defer(); err != nil {
    log.Println(err)
    return err
  }
	ctx.SetEphemeral(true)
	args := ctx.Options()
	name := args.Get(0).UserValue(ctx)

	player, err := p.models.Players.GetByDiscordID(name.ID)
	if err != nil {
		discord.ErrorMessage(
			ctx,
			"Unable to find Player",
			"No known player of name "+name.Username,
		)
		return err
	}

	role, err := p.models.Roles.Get(player.RoleID)
	if err != nil {
		discord.ErrorMessage(ctx, "Unable to find Role", "No known role of name "+name.Username)
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
