package help

import (
	"github.com/bwmarrin/discordgo"
)

func adminInventoryEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title: "Inventory",
		Description: "Overall on how to manage and update player inventories using `/inv`. *Most* inv command will fall under the template of /inv [category] [add/remove/set] [thing]. This command can be used in a player's confessional channel or in a channel that has been whitelisted for inventory commands.\n " +
			"You can add a channel to the whitelist by using the `/whitelist` command. ** One key thing to note is if you are in a player's confessional channel, you may omit the `player` argument and it will default to the player who's confessional channel you are in.**",

		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "For status and effect fields, you an also include an optional duration `argument` within their add argument. This follows a simple string format of human-readable duration string (e.g., '1h30m') to schedule a status or effect to expire after a certain amount of time. If no duration is specified, the status or effect will be permanent." +
					"If you remove a status or effect that was scheduled to expire, it will be removed from the scheduler to avoid accidental future deletions." +
					"For sake of minimal overhead, If you wish to 'extend' the duration of something, it is recommended to remove it and re-add it with the new duration.",
			},
			{
				Value: "When adding or removing items to a player's inventory you may do so in either their confessional channel, or within a created whitelist channel. You can whitelist a channel by using the `/whitelist` command. If you are in a , you may omit the `player` argument and it will default to the player who's confessional channel you are in.",
			},
		},
	}

	return msg
}
