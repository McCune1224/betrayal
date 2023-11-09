package commands

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/zekrotja/ken"
)

type Kill struct {
	models data.Models
}

// Description implements ken.SlashCommand.
func (*Kill) Description() string {
	return "Kill a player"
}

var _ ken.SlashCommand = (*Kill)(nil)

func (k *Kill) SetModels(models data.Models) {
	k.models = models
}

// Name implements ken.SlashCommand.
func (*Kill) Name() string {
	return discord.DebugCmd + "kill"
}

// Options implements ken.SlashCommand.
func (*Kill) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "norm",
			Description: "Normal kill",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "location",
			Description: "set kill player list location",
			Options: []*discordgo.ApplicationCommandOption{
				discord.ChannelCommandArg(true),
			},
		},
	}
}

func (k *Kill) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "norm", Run: k.killNorm},
		ken.SubCommandHandler{Name: "location", Run: k.killLocation},
	)
}

// Run implements ken.SlashCommand.
func (k *Kill) killNorm(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAuthorizedError(ctx)
	}
	// type cast ctx to subcommand context
	inv, err := inventory.Fetch(ctx, k.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	err = k.models.Inventories.UpdateProperty(inv, "is_alive", false)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	err = inventory.UpdateInventoryMessage(ctx, inv)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	userId := inv.DiscordID
	// get user via discordgo
	user, err := ctx.GetSession().User(userId)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	invs, err := k.models.Inventories.GetAll()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	hitlist, err := k.models.Hitlists.Get()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	err = UpdateHitlistMesage(ctx, invs, hitlist)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	return discord.SuccessfulMessage(ctx, "Player Killed", fmt.Sprintf("%s is marked dead", user.Username))
}

func (k *Kill) killLocation(ctx ken.SubCommandContext) (err error) {
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAuthorizedError(ctx)
	}

	channel := ctx.Options().GetByName("channel").ChannelValue(ctx)
	hitlistCreateMsg := discordgo.MessageEmbed{
		Title: "THE HITLIST IS ON THE WAY",
	}
	pinMsg, err := ctx.GetSession().ChannelMessageSendEmbed(channel.ID, &hitlistCreateMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	invs, err := k.models.Inventories.GetAll()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	embd := HitListBuilder(invs, ctx.GetSession())
	msg, err := ctx.GetSession().ChannelMessageEditEmbed(channel.ID, pinMsg.ID, embd)
	if err != nil {
		log.Print(err)
		discord.AlexError(ctx)
	}

	hitlist := data.Hitlist{
		PinChannel: channel.ID,
		PinMessage: msg.ID,
	}
	_, err = k.models.Hitlists.Insert(&hitlist)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	err = ctx.GetSession().ChannelMessagePin(channel.ID, msg.ID)
	if err != nil {
		log.Println(err)
		discord.AlexError(ctx)
	}
	return err
}

func HitListBuilder(invs []data.Inventory, s *discordgo.Session) *discordgo.MessageEmbed {
	humanReqTime := util.GetEstTimeStamp()
	fields := []*discordgo.MessageEmbedField{}
	deadTally := 0
	for _, inv := range invs {
		name, err := s.User(inv.DiscordID)
		nameStr := discord.MentionUser(name.ID)
		if err != nil {
			log.Println(err)
			return &discordgo.MessageEmbed{
				Title:       "Error getting user",
				Description: fmt.Sprintf("Failed getting %s inv", inv.DiscordID),
			}
		}
		if !inv.IsAlive {
			nameStr = fmt.Sprintf("%s %s", discord.EmojiDead, nameStr)
			deadTally += 1
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Value: nameStr,
		})
	}

	foot := &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Last updated: %s", humanReqTime),
	}

	aliveCount := len(invs) - deadTally
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "Alive",
		Value: fmt.Sprintf("%d", aliveCount),
	})
	message := discordgo.MessageEmbed{
		Title:       "Player List",
		Description: "Current status of each player",
		Fields:      fields,
		Color:       discord.ColorThemeBlack,
		Footer:      foot,
	}

	return &message
}

func UpdateHitlistMesage(ctx ken.Context, invs []data.Inventory, h *data.Hitlist) (err error) {
	sesh := ctx.GetSession()
	_, err = sesh.ChannelMessageEditEmbed(
		h.PinChannel,
		h.PinMessage,
		HitListBuilder(invs, sesh),
	)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Version implements ken.SlashCommand.
func (*Kill) Version() string {
	return "1.0.0"
}
