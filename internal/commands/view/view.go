package view

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zekrotja/ken"
)

type View struct {
	dbPool *pgxpool.Pool
}

// Description implements ken.SlashCommand.
func (v *View) Description() string {
	panic("unimplemented")
}

// Name implements ken.SlashCommand.
func (v *View) Name() string {
	panic("unimplemented")
}

// Options implements ken.SlashCommand.
func (v *View) Options() []*discordgo.ApplicationCommandOption {
	panic("unimplemented")
}

// Run implements ken.SlashCommand.
func (v *View) Run(ctx ken.Context) (err error) {
	panic("unimplemented")
}

// Version implements ken.SlashCommand.
func (v *View) Version() string {
	panic("unimplemented")
}

func (v *View) Initialize(pool *pgxpool.Pool) {
	v.dbPool = pool
}

var _ ken.SlashCommand = (*View)(nil)
