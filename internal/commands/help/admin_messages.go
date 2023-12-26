package help

import (
	"github.com/bwmarrin/discordgo"
)

func adminInventoryEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Inventory Admin Commands",
		Description: "Overall on how to manage and update player inventories using `/inv`. Biggest takeaway is that modifying a player's inventory generally follows the flow of `/inv [category] [add/remove/set] [player]`.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/inv [category] [add/remove/set] [player]`. This is how you add/remove/set items to a player's inventory. Can only be done in a player's confessional channel or a whitelisted channel. *(player argument is not mandatory if you're already in the player's confessional)*",
			},
			{
				Value: ">>> **status** and **effect** fields also include an optional argument `duration` when adding. It uses a string format of human-readable duration string (e.g., '1h30m') to schedule a status or effect to expire." +
					"For sake of minimal overhead, If you wish to 'extend' the duration of something, it is recommended to remove it and re-add it with the new duration.",
			},
			{
				Value: "`/inv create [target]`. Create an inventory for desired player (this command should be issued within the player's confessional as it will pin the inventory to this channel).",
			},
			{
				Value: "`/inv delete [target]`. Delete an inventory for desired player. Will also delete the pinned inventory within the player's confessional channel if it exists.",
			},
			{
				Value: "`/inv whitelist [add/remove] [channel]`. Add or Remove from the Whitelist for a channel for allowing inventory commands. This will allow you to issue inventory commands in the channel." +
					"You can verify which channels are whitelisted by using `/inv whitelist list`.",
			},
		},
	}

	return msg
}

func adminAllianceEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Alliance Admin Commands",
		Description: "All admin based commands will follow the flow of `/alliance admin [command] [args]`. Most of the admin commands are just here to approve requests put in by players.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Value: `/alliance admin pending, list all pending requests awaiting approval.`,
			},
			{
				Value: "`/alliance admin create [alliance name]`, approve creation of a new alliance. This will create a new alliance with the given name and auto make a channel and insert the player who put in the request.",
			},
			{
				Value: "`/alliance admin decline [alliance name]`, decline creation of a new alliance. This will delete the request and notify the player who put in the request.",
			},
			{
				Value: "`/alliance admin invite [player] [alliance channel]. Approve an invite request. This will add the player to the alliance, channel, and notify them in their confessional.",
			},
			{
				Value: "`/alliance admin wipe [alliance name]`. Wipe an alliance. This will delete the alliance, channel, and all pending requests for the alliance. (This is a destructive action and cannot be undone.)",
			},
		},
	}
	return msg
}

func adminRollEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Roll Admin Commands",
		Description: "Roll is a helper command that allows you to roll game events as well as items/abilities on the fly with confirmable menus before sending.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/roll manual [category] [level] [player]`, allows simulation one-off rolls for a player. *(must use something like `/inv item/aa add` to add the item to the player's inventory)*",
			},
			{
				Value: "`/roll [care_package / power_drop / item_rain] [player]`, Allows to do an event roll for target player. Will give an option to accept/decline the outcome. Will inform player in their confessional if accepted and auto add to their inventory.",
			},
			{
				Value: "`/roll wheel`, Fun command that will spin a wheel and give you a random event for the day.",
			},
		},
	}
	return msg
}

func adminBuyEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Buy Admin Command",
		Description: "Buy is a helper command that allows you to buy an item on behalf of a player. This is useful for when a player is unable to use the `/buy` command themselves. This command follows the flow of `/buy [item] [player]`.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/buy [item] [player]`. This is how you buy an item on behalf of a player. This command can be issued anywhere you could a confessional or whitelisted channel.",
			},
			{
				Value: "Buy will fail if the player does not have enough coins to buy the item.",
			},
		},
	}

	return msg
}

func adminKillEmebd() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Kill/Revive Admin Commands",
		Description: "Kill and Revive are helper commands that allows you to mark players as alive/dead.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/kill [player]`. This is how you kill a player. This command can be issued anywhere you could a confessional or whitelisted channel.",
			},
			{
				Value: "`/revive [player]`. This is how you revive a player. This command can be issued anywhere you could a confessional or whitelisted channel.",
			},
			{
				Value: "`/kill location [channel]`. Set the location to show status board for players that are marked as dead. This board ideally should be put in a channel that is not accessible to all players.",
			},
		},
	}
	return msg
}

func adminSetupEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Setup Admin Command",
		Description: "Setup is a helper command that allows you to determine roles for game creation. This command will walk you through the process of setting up the game. This command follows the flow of `/setup [player count] [deceptionist count] [good count] [evil count]`.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/setup [player count] `. Help generate the role list pool you'd like for the game. By default it will assume all user's with a deceiptionist role are in the game. But if you'd like to change the value, you can include the additional argument `deceptionist count`.",
			},
		},
	}
	return msg
}
