package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/zekrotja/ken"
)

type Test struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

var _ ken.SlashCommand = (*Test)(nil)

// Initialize implements BetrayalCommand.
func (t *Test) Initialize(m data.Models, s *scheduler.BetrayalScheduler) {
	t.models = m
	t.scheduler = s
}

// Description implements ken.SlashCommand.
func (*Test) Description() string {
	return "Dev Sandbox for commands"
}

// Name implements ken.SlashCommand.
func (*Test) Name() string {
	return "t"
}

// Options implements ken.SlashCommand.
func (*Test) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "channel",
			Description: "Channel related commands",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "get",
					Description: "Get details of a channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "delete",
					Description: "Delete target channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.ChannelCommandArg(true),
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "create",
					Description: "create target channel",
					Options: []*discordgo.ApplicationCommandOption{
						discord.StringCommandArg("name", "Name of the channel to create", true),
					},
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (t *Test) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandGroup{Name: "channel", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "get", Run: t.getChannel},
			ken.SubCommandHandler{Name: "delete", Run: t.deleteChannel},
			ken.SubCommandHandler{Name: "create", Run: t.createChanenl},
		}},
	)
}

func (t *Test) getChannel(ctx ken.SubCommandContext) (err error) {
	channelArg := ctx.Options().GetByName("channel").ChannelValue(ctx)

	currentChannelPerimssions := channelArg.PermissionOverwrites

	fields := []*discordgo.MessageEmbedField{}

	for _, perm := range currentChannelPerimssions {
		title := fmt.Sprintf("%d - %s - %d", perm.Type, perm.ID, perm.Deny)
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  title,
			Value: fmt.Sprintf("PERM TYPE: %d", perm.Allow),
		})
	}

	msg := &discordgo.MessageEmbed{
		Title:       channelArg.Name,
		Description: channelArg.ID,
		Fields:      fields,
	}

	return ctx.RespondEmbed(msg)
}

func (t *Test) deleteChannel(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	s := ctx.GetSession()

	targetChannel := ctx.Options().GetByName("channel").ChannelValue(ctx)
	log.Println(targetChannel)

	removed, err := s.ChannelDelete(targetChannel.ID)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to delete channel")
	}

	return discord.SuccessfulMessage(ctx, "Deleted Alliance channel", removed.Name)
}

// Version implements ken.SlashCommand.
func (*Test) Version() string {
	return "1.0.0"
}

func (t *Test) createChanenl(ctx ken.SubCommandContext) (err error) {
	// if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
	// 	return discord.NotAdminError(ctx)
	// }
	e := ctx.GetEvent()
	s := ctx.GetSession()
	if e.Member.User.ID != discord.McKusaID {
		return discord.NotAdminError(ctx)
	}
	channelArg := ctx.Options().GetByName("name").StringValue()
	channel, err := discord.CreateHiddenChannel(s, e, channelArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to create channel")
	}

	msg := &discordgo.MessageEmbed{}
	msg.Title = "Created Channel"
	msg.Description = fmt.Sprintf("Created channel %s", channel.Name)
	msg.Fields = []*discordgo.MessageEmbedField{}
	msg.Fields = append(msg.Fields, &discordgo.MessageEmbedField{
		Name:  "Channel ID",
		Value: channel.ID,
	})

	return ctx.RespondEmbed(msg)
}
