package help

import "github.com/bwmarrin/discordgo"

func actionHelpEmbed() *discordgo.MessageEmbed {
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

func InventoryHelpEmbed() *discordgo.MessageEmbed {
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

func AllianceHelpEmbed() *discordgo.MessageEmbed {
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
