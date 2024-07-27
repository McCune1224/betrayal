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
func (c *Channel) Version() string {
	return "1.0.0"
}

func (c *Channel) Initialize(pool *pgxpool.Pool) {
	c.dbPool = pool
}

var _ ken.SlashCommand = (*Channel)(nil)

// Description implements ken.SlashCommand.
func (c *Channel) Description() string {
	return "1.0.0"
}

// Name implements ken.SlashCommand.
func (c *Channel) Name() string {
	return "channel"
}

// Options implements ken.SlashCommand.
func (c *Channel) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "confessionals",
			Description: "View current confessional channel details",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
		c.adminCommandArgBuilder(),
		c.voteCommandArgBuilder(),
		c.actionCommandArgBuilder(),
	}
}

// Run implements ken.SlashCommand.
func (c *Channel) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		c.voteCommandGroupBuilder(),
		c.adminCommandGroupBuilder(),
		c.actionCommandGroupBuilder(),
		ken.SubCommandHandler{Name: "confessionals", Run: c.viewConfessionals},
	)
}

func (c *Channel) viewConfessionals(ctx ken.SubCommandContext) (err error) {
	q := models.New(c.dbPool)
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
