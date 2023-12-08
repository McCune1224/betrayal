package help

import "github.com/bwmarrin/discordgo"

func InventoryHelpEmbed() *discordgo.MessageEmbed {
	msg := &discordgo.MessageEmbed{
		Title:       "Inventory",
		Description: "**TLDR: Your bread and butter command as a command is going to be `/inv get`. and you can always view your inventory in your pinned confessional.**\n\n The inventory commands are used to manage your inventory. As a player, you can view your inventory to see a variety of information.\nYour inventory is updated in real time, so you can always see the most up-to-date information about your character.",
		Fields: []*discordgo.MessageEmbedField{
			{
				Value: "To get your inventory, use `/inv get [your name]`.",
			},
			{
				Value: "Your inventory is a collection of items, abilities, and other information that you have collected throughout the game. You can view your inventory at any time by using `/inv get [your name]`." +
					"Keep in mind that you can only view your own inventory.",
			},
		},

		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://media.giphy.com/media/l378xcbxNV5QYfygg/giphy.gif",
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
				Value: "To invite a player to your alliance, use `/alliance invite [player name]` a player will be added to your alliance channel *IF* they accept. (you must already be in an alliance for this to work).",
			},
			{
				Value: "To accept an alliance invite, use `/alliance accept [alliance name]`. Once an admin approves, you will be added to the alliance channel.",
			},
			{
				Value: "To leave an alliance, use `/alliance leave [alliance name]`. You will be removed from the alliance channel.",
			},
		},
	}
	return msg
}
