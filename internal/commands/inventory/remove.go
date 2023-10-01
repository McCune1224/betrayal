package inventory

import (
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/zekrotja/ken"
)

func (i *Inventory) removeAbility(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) removePerk(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) removeItem(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) removeStatus(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) removeImmunity(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) removeCoins(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) removeWhitelist(ctx ken.SubCommandContext) (err error) {
	ctx.SetEphemeral(true)
	channelArg := ctx.Options().GetByName("channel").ChannelValue(ctx)

	whitelists, _ := i.models.Whitelists.GetAll()
	if len(whitelists) == 0 {
		err = discord.SendSilentError(ctx, "No whitelisted channels", "Nothing here...")
		return err
	}

	for _, v := range whitelists {
		if v.ChannelID == channelArg.ID {
			i.models.Whitelists.Delete(v)
		}
		err = ctx.RespondMessage("Channel removed from whitelist.")
		return err
	}

	err = discord.SendSilentError(ctx, "Channel not found", "This channel is not whitelisted.")
	return err

}
