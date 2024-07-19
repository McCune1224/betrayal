package channels

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Channel struct {
	dbPool *pgxpool.Pool
}

// Version implements ken.SlashCommand.
func (b *Channel) Version() string {
	return "1.0.0"
}

func (b *Channel) Initialize(pool *pgxpool.Pool) {
	b.dbPool = pool
}

var _ ken.SlashCommand = (*Channel)(nil)

// Description implements ken.SlashCommand.
func (b *Channel) Description() string {
	return "1.0.0"
}

// Name implements ken.SlashCommand.
func (b *Channel) Name() string {
	return "channel"
}

// Options implements ken.SlashCommand.
func (b *Channel) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "vote",
			Description: "View / Update current voting channel details",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "create",
					Description: "Create the voting channel, this will replace any existing vote channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "action",
			Description: "View / Update current action channel details",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "create",
					Description: "Create the voting channel, this will replace any existing vote channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "confessionals",
			Description: "View current confessional channel details",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "admin",
			Description: "View / Update current admin channel details (needed for inventory commands)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Add a channel to allow inventory commands to run in",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "view",
					Description: "View current admin channel details",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "delete",
					Description: "Delete a channel from the list of allowed channels",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (b *Channel) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		// ken.SubCommandGroup{Name: "vote", SubHandler: []ken.CommandHandler{
		// 	ken.SubCommandHandler{Name: "create", Run: b.createVote},
		// 	ken.SubCommandHandler{Name: "view", Run: b.viewVote},
		// }},
		// ken.SubCommandGroup{Name: "action", SubHandler: []ken.CommandHandler{
		// 	ken.SubCommandHandler{Name: "create", Run: b.createAction},
		// 	ken.SubCommandHandler{Name: "view", Run: b.viewAction},
		// }},
		// ken.SubCommandGroup{Name: "admin", SubHandler: []ken.CommandHandler{
		// 	ken.SubCommandHandler{Name: "add", Run: b.addAdmin},
		// 	ken.SubCommandHandler{Name: "view", Run: b.viewAdmin},
		// 	ken.SubCommandHandler{Name: "delete", Run: b.deleteAdmin},
		// }},
		ken.SubCommandHandler{Name: "confessionals", Run: b.viewConfessionals},
	)
}

func (b *Channel) viewConfessionals(ctx ken.SubCommandContext) (err error) {
	q := models.New(b.dbPool)
	dbCtx := context.Background()

	currentConfessionals, err := q.ListPlayerConfessional(dbCtx)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}

	msg := ""
	for _, confessional := range currentConfessionals {
		confID := util.Itoa64(confessional.ChannelID)
		playerID := util.Itoa64(confessional.PlayerID)
		msg += fmt.Sprintf("%s - %s\n", discord.MentionUser(playerID), discord.MentionChannel(confID))
	}

	return ctx.RespondEmbed(&discordgo.MessageEmbed{
		Title:       "Currently Available Confessionals",
		Description: msg,
	})
}
