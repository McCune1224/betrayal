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
				Name:  "Basic Inventory Modifications",
				Value: "`/inv [category] [add/remove/set] [player]`. This is how you add/remove/set items to a player's inventory. Can only be done in a player's confessional channel or a whitelisted channel. *(player argument is not mandatory if you're already in the player's confessional)*",
			},
			{
				Name:  "Categories",
				Value: "Available categories: **ability**, **item**, **coin**, **status**, **perk**, **alignment**, **role**, **immunity**, **luck**, **death**, **notes**.",
			},
			{
				Name:  "Timed Effects (Status & Perks)",
				Value: "**status** and **perk** fields support an optional `duration` argument when adding. Use human-readable format (e.g., `1h30m`, `45m`, `2h`). This schedules automatic expiration. To extend a duration, remove and re-add with the new duration.",
			},
			{
				Name:  "Duration Examples",
				Value: "`/inv status add [player] poison 30m` (expires in 30 minutes)\n`/inv status add [player] bleeding 2h` (expires in 2 hours)\n`/inv perk add [player] shield 1h30m` (expires in 1.5 hours)",
			},
			{
				Name:  "Inventory Creation & Deletion",
				Value: "`/inv create [role] [player]` - Create inventory for a player (run in their confessional to auto-pin). `/inv delete [player]` - Delete inventory and remove pinned message.",
			},
			{
				Name:  "Whitelist Management",
				Value: "`/inv whitelist [add/remove] [channel]` - Add or remove a channel from the whitelist for inventory commands. `/inv whitelist list` - View all whitelisted channels. Whitelisted channels allow inventory modifications outside confessionals.",
			},
		},
	}

	return msg
}

func adminAllianceEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Alliance Admin Commands",
		Description: "Manage alliance requests and approvals. All admin commands follow the flow of `/alliance admin [command] [args]`. Admins act on requests submitted by players.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "View Pending Requests",
				Value: "`/alliance admin pending` - View all pending alliance requests (creation and invite requests) awaiting approval.",
			},
			{
				Name:  "Approve Alliance Creation",
				Value: "`/alliance admin create [alliance name]` - Approve a player's request to create an alliance. This creates the alliance, generates a Discord channel, and adds the requesting player.",
			},
			{
				Name:  "Decline Alliance Creation",
				Value: "`/alliance admin decline [alliance name]` - Reject a player's alliance creation request. Deletes the request and notifies the player in their confessional.",
			},
			{
				Name:  "Approve Alliance Invite",
				Value: "`/alliance admin invite [player] [alliance channel]` - Approve a player's invite to join an alliance. Adds the player to the alliance and channel, then notifies them in their confessional.",
			},
			{
				Name:  "Wipe Alliance",
				Value: "`/alliance admin wipe [alliance name]` - Remove an alliance entirely. Deletes the alliance, channel, and all pending requests. **This is destructive and cannot be undone.**",
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

func adminChannelsEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Channel Admin Commands",
		Description: "Manage game channels including confessionals, admin channels, voting channels, action channels, and lifeboards. Use `/channel [subgroup] [command] [options]` to configure these channels.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Admin Channel",
				Value: "`/channel admin add [channel]` - Set a channel for admin operations (whitelisted for inventory commands).\n`/channel admin delete [channel]` - Remove admin channel designation.\n`/channel admin list` - View current admin channels.",
			},
			{
				Name:  "Vote Channels",
				Value: "`/channel vote add [channel]` - Add a voting channel where players can submit votes.\n`/channel vote delete [channel]` - Remove a voting channel.\n`/channel vote list` - View all voting channels.",
			},
			{
				Name:  "Action Channels",
				Value: "`/channel action add [channel]` - Add an action submission channel.\n`/channel action delete [channel]` - Remove an action channel.\n`/channel action list` - View all action channels.",
			},
			{
				Name:  "Lifeboard",
				Value: "`/channel lifeboard set [channel]` - Set the channel for displaying player status/lifeboard.\n`/channel lifeboard delete` - Remove lifeboard channel.\n`/channel lifeboard list` - View current lifeboard channel.",
			},
			{
				Name:  "Confessionals",
				Value: "`/channel confessionals` - View all current player confessional channels and their details.",
			},
		},
	}
	return msg
}

func adminCycleEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Cycle Admin Commands",
		Description: "Manage game phases and cycles. Control progression through Day/Night phases and elimination phases using `/cycle [command]`.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "View Current Phase",
				Value: "`/cycle current` - Display the current game phase (e.g., Day 1, Elimination 2). Useful for tracking game progress.",
			},
			{
				Name:  "Advance to Next Phase",
				Value: "`/cycle next` - Automatically advance to the next phase. Day phases progress to Elimination phases and vice versa (Day 1 → Elimination 1 → Day 2).",
			},
			{
				Name:  "Manually Set Phase",
				Value: "`/cycle set [phase] [number]` - Manually override the current phase. Phase options: **Day** or **Elimination**. Example: `/cycle set Day 3` sets the game to Day 3.",
			},
			{
				Name:  "Broadcasting",
				Value: "When advancing to a new cycle, the game automatically broadcasts the new phase to all player confessionals, alliance channels, and funnel channels.",
			},
		},
	}
	return msg
}
