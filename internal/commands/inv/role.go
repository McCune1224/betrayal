package inv

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/inventory"
	"github.com/zekrotja/ken"
)

func (i *Inv) roleCommandGroupBuilder() ken.SubCommandGroup {
	return ken.SubCommandGroup{Name: "role", SubHandler: []ken.CommandHandler{
		ken.SubCommandHandler{Name: "set", Run: i.setRole},
	}}
}

func (i *Inv) roleCommandArgBuilder() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        "role",
		Description: "Change the player's role",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "the role to set them to",
				Options: []*discordgo.ApplicationCommandOption{
					discord.StringCommandArg("role", "name of the role", true),
					discord.UserCommandArg(false),
				},
			},
		},
	}
}

func (i *Inv) setRole(ctx ken.SubCommandContext) (err error) {
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
	roleNameArg := ctx.Options().GetByName("role").StringValue()
	q := models.New(i.dbPool)
	newRole, err := q.GetRoleByFuzzy(context.Background(), roleNameArg)
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}

	_, err = q.UpdatePlayerRole(context.Background(), models.
		UpdatePlayerRoleParams{
		ID:     h.GetPlayer().ID,
		RoleID: pgtype.Int4{Int32: newRole.ID, Valid: true},
	})
	if err != nil {
		log.Println(err)
		return discord.AlexError(ctx, "")
	}

	return discord.SuccessfulMessage(ctx, "Role Updated", fmt.Sprintf("Role has been updated to %s", newRole.Name))
}
