package help

import "github.com/bwmarrin/discordgo"

func actionHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Action",
		Description: "**TLDR: Just use `/action request [thing]`.** \n\n Actions are sent to admins for approval, and can be written in free response. As a general rule of thumb, if its being used from your inventory, use `/action request [thing]`.",

		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "To place a request, use `/action request [thing]`. This is setup so that your action request will be sent to hosts for processing. You will be informed by a host/co-host on the results of the action.",
			},
			{
				Value: "For example, if you're playing the role Wizard and wish to cast Fireball. You would do so by using `/action request cast Fireball`. You can also add additional information to your request by using `/action request Fireball at [player name]`.",
			},
			{
				Value: "The same can be done with other categories such as perks, items, etc. For example, if you wanted to use the item 'Tip', you would simply do `/action request Tip`.",
			},
			{
				Value: "As a reminder, these only put in a request for an action. The action will not be performed until a host/co-host approves it. You will be notified in your confessional when your action is processed.",
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
				Value: "To get your own inventory, use `/inv me [your name]`. You technically can get yourself with `/inv get [your name]`, but `/inv me` is much easier.",
			},
			{
				Value: "Your inventory is a collection of items, abilities, and other information that you have collected throughout the game. You can view your inventory at any time by using `/inv get [your name]`." +
					"Keep in mind that you can only view your own inventory.",
			},
		},
	}
	return msg
}

func AllianceHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Alliances",
		Description: "Commands for creating, joining, and leaving alliances. Keep in mind these commands make requests for hosts to approve, so you may need to wait a little bit before your command is processed.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "To request to create an alliance, use `/alliance request [your alliance name]`. Once admin approved, you will be notified in your confessional and a channel will be created for your alliance.",
			},
			{
				Value: "To invite a player to your alliance, use `/alliance invite [player name]`. An invite will be sent to the player's confessional. If they accept, they will be added to the alliance channel after admin approval.",
			},
			{
				Value: "To accept an alliance invite, use `/alliance accept [alliance name]`. Once an admin approves, you will be added to the alliance channel.",
			},
			{
				Value: "To leave an alliance, use `/alliance leave [alliance name]`. You will be removed from the alliance automatically channel.",
			},
		},
	}
	return msg
}
