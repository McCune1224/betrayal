package help

import (
	"github.com/bwmarrin/discordgo"
)

func adminInventoryEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Inventory",
		Description: "Overall on how to manage and update player inventories using `/inv`. Biggest takeaway is that modifying a player's inventory generally follows the flow of `/inv [category] [add/remove/set] [player]`.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/inv [category] [add/remove/set] [player]`. This is how you add/remove/set items to a player's inventory. Can only be done in a player's confessional channel or a whitelisted channel. *(player argument is not mandatory if you're already in the player's confessional)*",
			},
			{
				Value: ">>> **status** and **effect** fields also include an optional argument `duration` when adding. It uses a string format of human-readable duration string (e.g., '1h30m') to schedule a status or effect to expire." +
					"For sake of minimal overhead, If you wish to 'extend' the duration of something, it is recommended to remove it and re-add it with the new duration.",
			},
			{},
			{
				Value: "`/inv create [target]`. Create an inventory for desired player (this command should be issued within the player's confessional as it will pin the inventory to this channel).",
			},
			{
				Value: "`/inv delete [target]`. Delete an inventory for desired player. Will also delete the pinned inventory within the player's confessional channel if it exists.",
			},
			{
				Value: "`/inv whitelist [add/remove]`. Add or Remove from the Whitelist for a channel for allowing inventory commands. This will allow you to issue inventory commands in the channel.",
			},
		},
	}

	return msg
}
