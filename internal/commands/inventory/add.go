package inventory

import (
	"github.com/zekrotja/ken"
)

func (i *Inventory) addAbility(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addPerk(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addItem(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addStatus(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addImmunity(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}

func (i *Inventory) addCoins(ctx ken.SubCommandContext) (err error) {
	if err = ctx.Defer(); err != nil {
		return err
	}
	ctx.SetEphemeral(true)
	err = ctx.RespondMessage("Command not implemented go bug Alex")
	ctx.SetEphemeral(false)
	return err
}
