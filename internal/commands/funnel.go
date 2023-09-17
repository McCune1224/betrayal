package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

// Alex-action-funnel (TODO: make this reassignable/modular)
const funnelChannelID = "1144040897617612920"

type ActionFunnel struct{}

var _ ken.SlashCommand = (*ActionFunnel)(nil)

// Description implements ken.SlashCommand.
func (*ActionFunnel) Description() string {
	panic("unimplemented")
}

// Name implements ken.SlashCommand.
func (*ActionFunnel) Name() string {
	panic("unimplemented")
}

// Options implements ken.SlashCommand.
func (*ActionFunnel) Options() []*discordgo.ApplicationCommandOption {
	panic("unimplemented")
}

// Run implements ken.SlashCommand.
func (*ActionFunnel) Run(ctx ken.Context) (err error) {
	panic("unimplemented")
}

// Version implements ken.SlashCommand.
func (*ActionFunnel) Version() string {
	panic("unimplemented")
}
