package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
)

// Wrapped for handling creation of components with ken
type KenComponent struct {
	// How the button will look and be interacted with
	Button discordgo.MessageComponent
	// Actual logic for the button
	Handler func(ctx ken.ComponentContext) bool
}

// I'm not sure whats going on and at this point I'm too afraid to ask
func NewKenComponent(
	btn discordgo.Button,
	handler func(ctx ken.ComponentContext) bool,
) *KenComponent {
	return &KenComponent{}
}
