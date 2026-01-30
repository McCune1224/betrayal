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
		Description: "Setup is a helper command that assists with generating the optimal role distribution for game creation based on player count and team composition.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Basic Usage",
				Value: "`/setup [player_count]` or `/setup [player_count] [decept_count]`\n\nThe command generates a balanced role pool from the available roles. By default, it counts all users with a Deceptionist role in the guild, but you can override this.",
			},
			{
				Name:  "Parameters",
				Value: "**player_count** (required) - Total number of players participating in the game\n**decept_count** (optional) - Number of Deceptionists to include. If not specified, defaults to all guild members with the Deceptionist role.",
			},
			{
				Name:  "Role Distribution",
				Value: "The command generates a balanced pool with:\n• **Good roles** - Helpful to the town, work together\n• **Neutral roles** - Neither good nor evil, have own agendas\n• **Evil roles** - Deceptive roles working against the town\n\nDistribution is randomized to ensure variety and prevent predictability.",
			},
			{
				Name:  "Example Workflows",
				Value: "**Example 1:** `/setup 10` - Generate roles for 10 players using all current Deceptionists\n**Example 2:** `/setup 15 3` - Generate roles for 15 players with exactly 3 Deceptionists\n**Example 3:** `/setup 8 2` - Generate roles for 8 players with 2 Deceptionists",
			},
			{
				Name:  "What You Get",
				Value: "The command displays:\n• A list of recommended roles for each alignment (Good/Neutral/Evil)\n• Role count breakdown\n• A summary showing the balance of the game\n\nUse this as your basis for assigning players to roles.",
			},
			{
				Name:  "Tips & Best Practices",
				Value: "• Run this **before** the game starts to determine your role lineup\n• Ensure your Deceptionist count makes sense for your player pool (usually 15-25% of players)\n• The randomization means each run generates different results - run multiple times to find a distribution you like\n• Save the output or screenshot it before assigning roles to players",
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

func adminHealthcheckEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Healthcheck Admin Command",
		Description: "Verify your game setup before starting a new game. The healthcheck command displays all configured channels and player status at a glance.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Basic Usage",
				Value: "`/healthcheck` - Run this command to see the current state of all required and optional game channels.",
			},
			{
				Name:  "What It Checks",
				Value: "• **Admin Channels** (required) - Channels for inventory management\n• **Vote Channel** (required) - Channel for vote submissions\n• **Action Channel** (required) - Channel for action submissions\n• **Lifeboard** (optional) - Channel showing player status board\n• **Player Confessionals** (optional) - List of all player confessional channels\n• **Players** - Count of alive and dead players\n• **Game Cycle** - Current phase (Day/Elimination number)",
			},
			{
				Name:  "Status Indicators",
				Value: "✅ Green checkmarks indicate configured and ready\n⚠️ Warning icons indicate optional features not yet configured\n❌ Red X icons indicate missing required channels (cannot start game)",
			},
			{
				Name:  "When to Run",
				Value: "Run this command **before starting a new game** to ensure all required channels are configured. All three required channels (Admin, Vote, Action) must be set before you can proceed.",
			},
			{
				Name:  "Setup Instructions",
				Value: "If channels are missing, use `/channel admin`, `/channel vote`, or `/channel action` to configure them. See `/help admin channels` for detailed setup instructions.",
			},
		},
	}

	return msg
}

func adminTarotEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Tarot Admin Commands",
		Description: "Configure and reset tarot modes that keep state.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Reset In-Memory State",
				Value: "`/tarot reset scope:per_user user:@X` – clear a specific user’s assignment\n`/tarot reset scope:per_user` – clear all per-user assignments for this guild\n`/tarot reset scope:guild_deck` – reshuffle the guild deck (start fresh)\n`/tarot reset scope:all` – clear all tarot state",
			},
			{
				Name:  "Modes Overview",
				Value: "`deterministic` – stable by guild+user\n`random` – fresh draw each time\n`per_user` – remembers a card per user (in-memory)\n`guild_deck` – deals without replacement per guild",
			},
			{
				Name:  "Persistence",
				Value: "By default, tarot state is in-memory and resets on restart. Ask the devs to enable DB persistence if you want state to survive restarts.",
			},
		},
	}
	return msg
}
