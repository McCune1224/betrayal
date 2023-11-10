package commands

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/commands/inventory"
	"github.com/mccune1224/betrayal/internal/cron"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

type Revive struct {
	models    data.Models
	scheduler *cron.BetrayalScheduler
}

func (r *Revive) Initialize(models data.Models, scheduler *cron.BetrayalScheduler) {
	r.models = models
	r.scheduler = scheduler
}

var _ ken.SlashCommand = (*Revive)(nil)

// Description implements ken.SlashCommand.
func (*Revive) Description() string {
	return "Revive a player"
}

// Name implements ken.SlashCommand.
func (*Revive) Name() string {
	return discord.DebugCmd + "revive"
}

// Options implements ken.SlashCommand.
func (*Revive) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "norm",
			Description: "Normal revive",
			Options: []*discordgo.ApplicationCommandOption{
				discord.UserCommandArg(true),
			},
		},
	}
}

// Run implements ken.SlashCommand.
func (r *Revive) Run(ctx ken.Context) (err error) {
	return ctx.HandleSubCommands(
		ken.SubCommandHandler{Name: "norm", Run: r.normRevive},
	)
}

func (r *Revive) normRevive(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	// type cast ctx to subcommand context
	inv, err := inventory.Fetch(ctx, r.models, true)
	if err != nil {
		if errors.Is(err, inventory.ErrNotAuthorized) {
			return discord.NotAuthorizedError(ctx)
		}
		return discord.ErrorMessage(ctx, "Failed to find inventory.", "If not in confessional, please specify a user")
	}
	inv.IsAlive = true

	err = r.models.Inventories.UpdateProperty(inv, "is_alive", true)
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
	invs, err := r.models.Inventories.GetAll()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}
	hitlist, err := r.models.Hitlists.Get()
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	err = UpdateHitlistMesage(ctx, invs, hitlist)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx)
	}

	return discord.SuccessfulMessage(ctx, "Revived", fmt.Sprintf("%s is marked alive", user.Username))
}

// Version implements ken.SlashCommand.
func (*Revive) Version() string {
	return "1.0.0"
}
