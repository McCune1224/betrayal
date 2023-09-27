package inventory

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
	"github.com/zekrotja/ken/examples/middlewares/middlewares"
)

var optional = discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionBoolean,
	Name:        "hidden",
	Description: "Make view hidden or public (default hidden)",
	Required:    false,
}

type Inventory struct {
	models data.Models
}

var (
	_ ken.SlashCommand                = (*Inventory)(nil)
	_ middlewares.RequiresRoleCommand = (*Inventory)(nil)
)

// RequiresRole implements middlewares.RequiresRoleCommand.
func (*Inventory) RequiresRole() string {
	return "Bot Developer"
}

func (i *Inventory) Type() discordgo.ApplicationCommandType {
	return discordgo.ChatApplicationCommand
}

func (i *Inventory) SetModels(models data.Models) {
	i.models = models
}

// Description implements ken.SlashCommand.
func (*Inventory) Description() string {
	return "Command for managing inventory"
}

// Name implements ken.SlashCommand.
func (*Inventory) Name() string {
	return discord.DebugCmd + "inventory"
}

// Options implements ken.SlashCommand.
func (*Inventory) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "get",
			Description: "get player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "player",
					Description: "Player to get inventory for",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "create a new player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "player",
					Description: "Player to create inventory for",
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
			Name:        "update",
			Description: "(test) update player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "whomst",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "content",
					Description: "content",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "add",
			Description: "add to player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "ability",
					Description: "add an ability",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the ability",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "perk",
					Description: "add a perk",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the perk",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "item",
					Description: "add an item",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the item",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "add a status",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the status",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "immunity",
					Description: "add an immunity",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the immunity",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "remove",
			Description: "remove to player's inventory",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "ability",
					Description: "remove an ability",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the ability",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "perk",
					Description: "remove a perk",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the perk",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "item",
					Description: "remove an item",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the item",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "remove a status",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the status",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "immunity",
					Description: "remove an immunity",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Name of the immunity",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (i *Inventory) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "get", Run: i.get},
		ken.SubCommandHandler{Name: "create", Run: i.create},
		ken.SubCommandHandler{Name: "update", Run: i.update},
		ken.SubCommandGroup{Name: "add", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "ability", Run: i.addAbility},
			ken.SubCommandHandler{Name: "perk", Run: i.addPerk},
			ken.SubCommandHandler{Name: "item", Run: i.addItem},
			ken.SubCommandHandler{Name: "status", Run: i.addStatus},
			ken.SubCommandHandler{Name: "immunity", Run: i.addImmunity},
		}},
		ken.SubCommandGroup{Name: "remove", SubHandler: []ken.CommandHandler{
			ken.SubCommandHandler{Name: "ability", Run: i.removeAbility},
			ken.SubCommandHandler{Name: "perk", Run: i.removePerk},
			ken.SubCommandHandler{Name: "item", Run: i.removeItem},
			ken.SubCommandHandler{Name: "status", Run: i.removeStatus},
			ken.SubCommandHandler{Name: "immunity", Run: i.addImmunity},
		}},
	)
}

func (i *Inventory) get(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	player := ctx.Options().GetByName("player").UserValue(ctx)
	_, err = i.models.Inventories.GetByDiscordID(player.ID)
	if err != nil {
		ctx.SetEphemeral(true)
		ctx.RespondError(err.Error(), "Failed to get Player Inventory")
		ctx.SetEphemeral(false)
	}
	return err
}

func (i *Inventory) create(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}

	playerArg := ctx.Options().GetByName("player").UserValue(ctx)
	roleArg := ctx.Options().GetByName("role").StringValue()

	_, err = i.models.Roles.GetByName(roleArg)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to get Role", err.Error())
		return err
	}

	event := ctx.GetEvent()
	channelID := event.ChannelID

	respMsg := "Create Inventory - " + playerArg.Username
	pinMsg, err := ctx.GetSession().ChannelMessageSend(channelID, respMsg)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to send message", err.Error())
		return err
	}

	err = ctx.GetSession().ChannelMessagePin(channelID, pinMsg.ID)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to pin message", err.Error())
		return err
	}

	ctx.RespondMessage("Creating Player Inventory...")

	inv := &data.Inventory{
		DiscordID:      playerArg.ID,
		UserPinChannel: channelID,
		UserPinMessage: pinMsg.ID,
	}

	_, err = i.models.Inventories.Insert(inv)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to insert inventory", err.Error())
	}

	return nil
}

func (i *Inventory) update(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	user := ctx.Options().GetByName("user").UserValue(ctx)
	content := ctx.Options().GetByName("content").StringValue()

	inv, err := i.models.Inventories.GetByDiscordID(user.ID)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to get inventory", err.Error())
		return err
	}

	inv.Content = content
	err = i.models.Inventories.Update(inv)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to update inventory", err.Error())
		return err
	}

	sesh := ctx.GetSession()
	_, err = sesh.ChannelMessageEdit(inv.UserPinChannel, inv.UserPinMessage, content)
	if err != nil {
		discord.SendSilentError(ctx, "Failed to update message", err.Error())
		return err
	}

	ctx.SetEphemeral(true)
	ctx.RespondMessage("Updated Inventory message")
	ctx.SetEphemeral(false)

	return err
}

// Version implements ken.SlashCommand.
func (*Inventory) Version() string {
	return "1.0.0"
}
