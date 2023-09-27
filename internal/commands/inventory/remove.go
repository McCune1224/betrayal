package inventory

import "github.com/zekrotja/ken"

func (i *Inventory) removeAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) removePerk(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
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
