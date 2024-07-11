package inv

// func (i *Inv) immunityCommandGroupBuilder() ken.SubCommandGroup {
// 	return ken.SubCommandGroup{Name: "immunity", SubHandler: []ken.CommandHandler{
// 		ken.SubCommandHandler{Name: "add", Run: i.addImmunity},
// 		ken.SubCommandHandler{Name: "remove", Run: i.removeImmunity},
// 	}}
// }

// func (i *Inv) addImmunity(ctx ken.SubCommandContext) (err error) {
// 	if err = ctx.Defer(); err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
// 		return discord.NotAdminError(ctx)
// 	}
// 	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
// 	if err != nil {
// 		log.Println(err)
// 		return discord.AlexError(ctx, "failed to init inv handler")
// 	}
// 	defer h.UpdateInventoryMessage(ctx.GetSession())
// 	q := models.New(i.dbPool)
// 	existingImmunities, _ := q.ListPlayerImmunity(context.Background(), h.GetPlayer().ID)
// 	if len(existingImmunities) > 0 {
// 		for _, immunity := range existingImmunities {
// 			if immunity.Name == immunityArg {
// 			}
// 		}
// 	}
//
// }

// func (i *Inv) removeImmunity(ctx ken.SubCommandContext) (err error) {
// 	if err = ctx.Defer(); err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	if !discord.IsAdminRole(ctx, discord.AdminRoles...) {
// 		return discord.NotAdminError(ctx)
// 	}
// 	h, err := inventory.NewInventoryHandler(ctx, i.dbPool)
// 	if err != nil {
// 		log.Println(err)
// 		return discord.AlexError(ctx, "failed to init inv handler")
// 	}
// 	defer h.UpdateInventoryMessage(ctx.GetSession())
// }
