package inv

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inv) abilityCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "ability", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "add", Run: i.addAbility},
		ken.SubCommandHandler{Name: "delete", Run: i.deleteAbility},
		ken.SubCommandHandler{Name: "set", Run: i.setAbility},
	}}
}

func (i *Inv) abilityCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "ability",
		Description: "create/update/delete an ability in an inventory",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add an ability",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("ability", "Ability to add", true),
					discord.IntCommandArg("quantity", "amount of charges to add", false),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "delete",
				Description: "Delete an ability",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("ability", "Ability to add", true),
					discord.UserCommandArg(false),
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "set the quantity of an ability",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("ability", "Ability to add", true),
					discord.IntCommandArg("quantity", "amount of charges to st", true),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) addAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())

	abilityNameArg := ctx.Options().GetByName("ability").StringValue()
	// quantity := int32(ctx.Options().GetByName("quantity").IntValue())
	quantity := int32(1)
	if quantityArg, ok := ctx.Options().GetByNameOptional("quantity"); ok {
		quantity = int32(quantityArg.IntValue())
	}
	ability, err := h.AddAbility(abilityNameArg, quantity)
	if err != nil {
		if err.Error() == "ability already added" {
			ability, err := h.UpdateAbility(abilityNameArg, quantity)
			if err != nil {
				log.Println(err)
				return discord.AlexError(ctx, "failed to add ability")
			}
			return discord.SuccessfulMessage(ctx, "Ability Updated", fmt.Sprintf("Ability %s updated", ability.Name))
		}

		log.Println(err)
		return discord.AlexError(ctx, "failed to add ability")
	}
	return discord.SuccessfulMessage(ctx, "Ability Added", fmt.Sprintf("Added ability %s", ability.Name))
}

func (i *Inv) deleteAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())

	abilityNameArg := ctx.Options().GetByName("ability").StringValue()
	ability, err := h.RemoveAbility(abilityNameArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to remove ability")
	}
	return discord.SuccessfulMessage(ctx, "Ability Removed", fmt.Sprintf("Removed ability %s", ability.Name))
}

func (i *Inv) setAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		log.Println(err)
		return err
	}
	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
		return discord.NotAdminError(ctx)
	}
	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to init inv handler")
	}
	defer h.UpdateInventoryMessage(ctx.GetSession())

	abilityNameArg := ctx.Options().GetByName("ability").StringValue()
	quantityArg := int(ctx.Options().GetByName("quantity").IntValue())
	ability, err := h.UpdateAbility(abilityNameArg, int32(quantityArg))
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "failed to remove ability")
	}
	return discord.SuccessfulMessage(ctx, "Ability Updated Charges", fmt.Sprintf("ability %s set to %d", ability.Name, quantityArg))
}
