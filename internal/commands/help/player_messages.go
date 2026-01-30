package help

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
)

func playerActionHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Action",
		Description: "**TLDR: Just use `/action request [thing]` or `/action request [thing at person].** \n\n Actions are sent to admins for approval.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/action request [thing]` to request an action for processing. This is setup so that your action request will be sent to hosts for processing. You will be informed by a host/co-host on the results of the action either in the form of emoji reaction (confirming that it was received) or if feedback is applicable, through host feeback).",
			},
			{
				Value: "For example with an item, if you wanted to use the item 'Tip', you would simply do `/action request Tip`. No need to specify on action's that don't ask for a target.",
			},
			{
				Value: "For example with an ability, if you have the ability 'Disappear' (either from being the neutral role Ghost, or via AA), you can cast it by submitting `/action request cast Disappear`.",
			},
			{
				Value: "For example with an ability that requires a target, if you have the ability 'Fireball' and want to use it on player 'Greg', you would do `/action request Fireball at Greg`. (*For abilities that require a target, you **MUST** specify a player*)",
			},
			{
				Value: "*As a reminder, these only put in a request for an action. The action will not be performed until a host/co-host approves it.*",
			},
		},
	}
	return msg
}

func playerInventoryHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Inventory",
		Description: "**TLDR: Just use `/inv me`.** Your inventory is also updated in real time and is always pinned as a message in confessional channel.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/inv me [your name]` to get your own inventory. You technically can get yourself with `/inv get [your name]`, but `/inv me` is much easier.",
			},
			{
				Value: "Your inventory is a collection of items, abilities, and other information that you have collected throughout the game. You can view your inventory at any time by using `/inv get [your name]`." +
					"Keep in mind that you can only view your own inventory.",
			},
			{
				Value: "You'll notice that there are a lot of other inventory commands. These are for hosts to use to update your inventory. You will not need to utilize these commands and can be disregarded.",
			},
		},
	}
	return msg
}

func playerAllianceHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Alliances",
		Description: "`/alliance` allows you request creating, joining, and leaving alliances. Mostly everything alliance based needs admin approval. Once approved, you will be notified in your confessional.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/alliance request [your alliance name]` to request making a new alliance. Once admin approved, you will be notified in your confessional and a channel will be created for your alliance.",
			},
			{
				Value: "`/alliance invite [player name] [alliance name]` to invite a player. You must already be a member within the alliance to send the invite. An invite will be sent to the player's confessional. If they accept, they will be added to the alliance channel after admin approval.",
			},
			{
				Value: "`/alliance accept [alliance name]` to request accepting alliance invite. Once an admin approves, you will be added to the alliance channel.",
			},
			{
				Value: "`/alliance leave [alliance name]` to leave the alliance. You will be no longer be associated with the requested alliance and be removed from the alliance channel automatically.",
			},
		},
	}
	return msg
}

func playerListHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "List",
		Description: "**TLDR: Just use `/list [category] [thing]`.**List is a command that allows you to view a list of things. Some useful things to pull up here are things like active_role, events.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/list [category] [thing]` to view a list of things. For example, `/list active_role` will show you a list of all active roles.",
			},
			{
				Value: fmt.Sprintf("You will notice that there are pre-defined categories. These are the only categories that you can use with `/list`. Let %s know if you would like to see a new category added.", discord.MentionUser(discord.McKusaID)),
			},
		},
	}
	return msg
}

func playerVoteHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Vote",
		Description: "**TLDR: Just use `/vote player [target]` or `/vote batch [target1], [target2]...`**.Vote is a command that allows you to vote on who to eliminate for today's vote. You can choose to vote for one or many players depending on if you have an ability or item that allows you to do so.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/vote player [target]` to vote on a player. For example, `/vote player Greg` will vote for Greg.",
			},
			{
				Value: "`/vote batch [tagets]` to vote on multiple players. The targets is free form. Feel free to use commas, spaces, or whatever you want to separate the targets. For example, `/vote batch Greg, Bob, Joe` will vote for Greg, Bob, and Joe.",
			},
		},
	}
	return msg
}

func playerViewHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "View",
		Description: "**TLDR: Just use `/view [category] [thing]`.** View is a command that allows you to quickly get info on many roles, abilities, perks, items, etc. Some useful things to pull up here are things like abilities, and perks. If the ability/perk is associated with a role, you will be provided a button to view the role.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "`/view [category] [thing]` to view a thing. For example, `/view role Wizard` will give you a full description of the Wizard role, including its abilities, alignment, perks, and description. These can be used to help you infer current game info or perhaps help you with your action requests.",
			},
			{
				Value: fmt.Sprintf("You will notice that there are pre-defined categories. These are the only categories that you can use with `/view`.  Let %s know if you would like to see a new category added.", discord.MentionUser(discord.McKusaID)),
			},
		},
	}
	return msg
}

func playerSearchHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Search",
		Description: "**TLDR: Just use `/search [type] [keyword]`.** Search is a command that allows you to discover what abilities and items are available in the game. This is useful for strategizing and deducing what might be happening in the game based on what effects you see.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Ability Search",
				Value: "`/search ability [keyword]` to search for abilities by keyword in both name and description. For example, `/search ability fire` will show you all abilities with \"fire\" in the name or description. Results are paginated - use the **Next** and **Previous** buttons to browse through results. Each ability shows its rarity, default charges, and which role(s) have that ability, helping you gauge its power level and make strategic deductions.",
			},
			{
				Name:  "Understanding Ability Results",
				Value: "Abilities with a **\\*** next to their name may have been obtained via an **any-ability**. This means they could exist on any role, not just the role-specific ones listed. For a detailed explanation of any-abilities, see: https://discord.com/channels/1096058997477490861/1096400207807271033/1096405318835118270",
			},
			{
				Name:  "Item Search",
				Value: "`/search item [keyword]` to search for items by keyword in both name and description. For example, `/search item relic` will show you all items with \"relic\" in the name or description (like \"Cursed Relic Grenade\"). Items display their rarity and cost, which can help you understand the economy and what to look for on the market.",
			},
			{
				Name:  "Pro Tips",
				Value: "• By default, both ability and item searches look in **names and descriptions**. Add `exclude_title: true` to search only in descriptions (e.g., `/search ability heal exclude_title: true`)\n• Search for broad terms like \"damage\", \"protect\", \"heal\", \"vote\" to discover related abilities and items\n• Ability searches now show which role(s) have that ability, helping you deduce what role a player might be\n• Look for the **\\*** asterisk on ability names - it indicates an any-ability that could be on any role\n• Results are sorted by rarity, with rarest items first",
			},
			{
				Name:  "Strategic Use",
				Value: "Use `/search` to gather intel about what's available in the game. By understanding the pool of abilities and items, you can make better deductions about what other players might have. Combined with the role information shown for abilities, you can narrow down which roles are in play. The asterisk helps you identify abilities that blur the line between roles.",
			},
		},
	}
	return msg
}

func playerTarotHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Tarot",
		Description: "Draw tarot cards for fun and flavor. Modes let you choose how unique the draw should be.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Basic Usage",
				Value: "`/tarot draw` – draws a card deterministically based on your user and guild (stable result).",
			},
			{
				Name:  "Modes",
				Value: "`mode: deterministic` (default) – stable per user in this guild\n`mode: random` – fresh random draw each time\n`mode: per_user` – assigns and remembers a unique card for you (in-memory)\n`mode: guild_deck` – deals cards without replacement for the guild until reset",
			},
			{
				Name:  "Options",
				Value: "`reversed: true` – force reversed orientation",
			},
			{
				Name:  "Notes",
				Value: "Per-user and guild-deck modes reset when the bot restarts unless an admin uses persistence (not enabled by default).",
			},
		},
	}
	return msg
}
