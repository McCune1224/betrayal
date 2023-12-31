package commands

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/scheduler"
	"github.com/mccune1224/betrayal/internal/util"
	"github.com/mccune1224/betrayal/pkg/data"
	"github.com/zekrotja/ken"
)

type Kill struct {
	models    data.Models
	scheduler *scheduler.BetrayalScheduler
}

// Description implements ken.SlashCommand.
func (*Kill) Description() string {
	return "Kill a player"
}

var _ ken.SlashCommand = (*Kill)(nil)

func (k *Kill) Initialize(models data.Models, scheduler *scheduler.BetrayalScheduler) {
	k.models = models
	k.scheduler = scheduler
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
			Name:        "player",
			Description: "Mark a player as dead and update kill list.",
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
		ken.SubCommandHandler{Name: "player", Run: k.killNorm},
		ken.SubCommandHandler{Name: "location", Run: k.killLocation},
	)
}

// Run implements ken.SlashCommand.
func (k *Kill) killNorm(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	// type cast ctx to subcommand context
	inv, err := inventory.Fetch(ctx, k.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAdminError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	err = k.models.Inventories.UpdateProperty(inv, "is_alive", false)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update death status")
	}

	inv.IsAlive = false
	err = inventory.UpdateInventoryMessage(ctx.GetSession(), inv)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update inventory message")
	}

	userId := inv.DiscordID
	// get user via discordgo
	user, err := ctx.GetSession().User(userId)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get user")
	}

	invs, err := k.models.Inventories.GetAll()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get inventories")
	}

	hitlist, err := k.models.Hitlists.Get()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get kill list")
	}

	err = UpdateHitlistMesage(ctx, invs, hitlist)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update kill list")
	}

	return discord.SuccessfulMessage(ctx, "Player Killed", fmt.Sprintf("%s is marked dead", user.Username))
}

func (k *Kill) killLocation(ctx ken.SubCommandContext) (err error) {
	if err := ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}

	channel := ctx.Options().GetByName("channel").ChannelValue(ctx)
	hitlistCreateMsg := discordgo.MessageEmbed{
		Title: "Building Player Status Board",
	}

	pinMsg, err := ctx.GetSession().ChannelMessageSendEmbed(channel.ID, &hitlistCreateMsg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to send message")
	}
	invs, err := k.models.Inventories.GetAll()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to get inventories")
	}

	embd := HitListBuilder(invs, ctx.GetSession())

	msg, err := ctx.GetSession().ChannelMessageEditEmbed(channel.ID, pinMsg.ID, embd)
	if err != nil {
		log.Print(err)
		discord.AlexError(ctx, "Failed to edit message")
	}

	currHitlist, _ := k.models.Hitlists.Get()
	if currHitlist != nil {
		// Incase there is already a hitlist, delete the currently pinned message
		err = ctx.GetSession().ChannelMessageDelete(currHitlist.PinChannel, currHitlist.PinMessage)
		if err != nil {
			log.Println(err)
			return discord.ErrorMessage(ctx, "Failed to delete old hitlist", "Please delete the old hitlist manually and try again")
		}
	}

	hitlist := data.Hitlist{
		PinChannel: channel.ID,
		PinMessage: msg.ID,
	}
	_, err = k.models.Hitlists.Upsert(&hitlist)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "Failed to update kill list")
	}
	err = ctx.GetSession().ChannelMessagePin(channel.ID, msg.ID)
	if err != nil {
		log.Println(err)
		discord.AlexError(ctx, "Failed to pin message")
	}
	return discord.SuccessfulMessage(ctx, "Hitlist Location Set", fmt.Sprintf("Hitlist location set to %s", channel.Mention()))
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
