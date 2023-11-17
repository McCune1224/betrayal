package scheduler

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/data"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/internal/util"
)

func (bs *BetrayalScheduler) ScheduleEffect(effect string, inv *data.Inventory, duration time.Duration, job interface{}) error {
	jobData := data.InventoryCronJob{
		InventoryID:       inv.UserPinChannel,
		InventoryCategory: "effect",
		InventoryAction:   "remove",
		InventoryValue:    effect,
		StartTime:         time.Now().Unix(),
		InvokeTime:        time.Now().Add(duration).Unix(),
	}
	jobID := jobData.GenerateJobID()
	jobData.JobID = jobID
	err := bs.dbJobs.InventoryCronJobs.Insert(&jobData)
	if err != nil {
		return err
	}
	err = bs.UpsertJob(jobID, duration, job)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BetrayalScheduler) QueueScheduleJobs(session *discordgo.Session) error {
	bs.cleanup()
	jobs, err := bs.dbJobs.InventoryCronJobs.GetByCategory("effect")
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Found %d jobs to queue", len(jobs))
	for _, job := range jobs {
		inventory, err := bs.dbJobs.Inventories.GetByPinChannel(job.InventoryID)
		if err != nil {
			log.Printf("Failed to find inventory for %s", job.InventoryID)
			continue
		}
		jobID := job.GenerateJobID()
		log.Println(time.Now().Unix() > job.InvokeTime)
		if time.Now().Unix() > job.InvokeTime {
			log.Println("DELETING JOB ID", jobID)
			err := bs.DeleteJob(jobID)
			if err != nil {
				log.Println("FAILED TO DELETE JOB", err)
			}
			continue
		}
		err = bs.UpsertJob(jobID, time.Duration(job.InvokeTime-job.StartTime)*time.Second, func() {
			// check to make sure the duration is still valid (check by making sure the start time + the duration is still in the future)
			err = bs.ScheduleEffect(job.InventoryValue, inventory, time.Duration(job.InvokeTime), func() {
				for k, v := range inventory.Effects {
					if strings.EqualFold(job.InventoryValue, v) {
						inventory.Effects = append(inventory.Effects[:k], inventory.Effects[k+1:]...)
						err = bs.dbJobs.Inventories.UpdateEffects(inventory)
						if err != nil {
							log.Println(err)
							msg := &discordgo.MessageEmbed{
								Title:       "Failed to remove effect",
								Description: fmt.Sprintf("Failed to remove effect %s from inventory", job.InventoryValue),
								Fields: []*discordgo.MessageEmbedField{
									{
										Name:  "Error",
										Value: err.Error(),
									},
								},
							}
							_, err := session.ChannelMessageSendEmbed(job.InventoryID, msg)
							if err != nil {
								log.Println(err)
								return
							}
						}

						return
					}
				}

				msg := discordgo.MessageEmbed{
					Title:       "Effect Expired",
					Description: fmt.Sprintf("Effect %s has expired", job.InventoryValue),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Target Inventory",
							Value: fmt.Sprintf("<@%s>", inventory.DiscordID),
						},
						{
							Name:  "Target Channel",
							Value: fmt.Sprintf("<#%s>", inventory.UserPinChannel),
						},
					},
					Color:     0x00ff00,
					Timestamp: time.Now().Format(time.RFC3339),
				}
				_, err := session.ChannelMessageSendEmbed(job.InventoryID, &msg)
				if err != nil {
					log.Println(err)
				}
				updateInventoryMessage(session, inventory)
			})
		})
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("Queued job %s to execute at %s", jobID, util.GetEstTimeStampFromDuration(time.Duration(job.InvokeTime-job.StartTime)*time.Second))
	}
	return nil
}

// FIXME: Have to manually copy this command because import cycle issues
func updateInventoryMessage(sesh *discordgo.Session, i *data.Inventory) (err error) {
	_, err = sesh.ChannelMessageEditEmbed(
		i.UserPinChannel,
		i.UserPinMessage,
		inventoryEmbedBuilder(i, false),
	)
	if err != nil {
		return err
	}
	return nil
}

// FIXME: Have to manually copy this command because import cycle issues
func inventoryEmbedBuilder(
	inv *data.Inventory,
	host bool,
) *discordgo.MessageEmbed {
	roleField := &discordgo.MessageEmbedField{
		Name:   "Role",
		Value:  inv.RoleName,
		Inline: true,
	}
	alignmentEmoji := discord.EmojiAlignment
	alignmentField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Alignment", alignmentEmoji),
		Value:  inv.Alignment,
		Inline: true,
	}

	// show coin bonus x100
	cb := inv.CoinBonus * 100
	coinStr := fmt.Sprintf("%d", inv.Coins) + " [" + fmt.Sprintf("%.2f", cb) + "%]"
	coinField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Coins", discord.EmojiCoins),
		Value:  coinStr,
		Inline: true,
	}
	abilitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Abilities", discord.EmojiAbility),
		Value:  strings.Join(inv.Abilities, "\n"),
		Inline: true,
	}
	perksField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Perks", discord.EmojiPerk),
		Value:  strings.Join(inv.Perks, "\n"),
		Inline: true,
	}
	anyAbilitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Any Abilities", discord.EmojiAnyAbility),
		Value:  strings.Join(inv.AnyAbilities, "\n"),
		Inline: true,
	}
	itemsField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Items (%d/%d)", discord.EmojiItem, len(inv.Items), inv.ItemLimit),
		Value:  strings.Join(inv.Items, "\n"),
		Inline: true,
	}
	statusesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Statuses", discord.EmojiStatus),
		Value:  strings.Join(inv.Statuses, "\n"),
		Inline: true,
	}

	immunitiesField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Immunities", discord.EmojiImmunity),
		Value:  strings.Join(inv.Immunities, "\n"),
		Inline: true,
	}
	effectsField := &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Effects", discord.EmojiEffect),
		Value:  strings.Join(inv.Effects, "\n"),
		Inline: true,
	}
	isAlive := ""
	if inv.IsAlive {
		isAlive = fmt.Sprintf("%s Alive", discord.EmojiAlive)
	} else {
		isAlive = fmt.Sprintf("%s Dead", discord.EmojiDead)
	}

	deadField := &discordgo.MessageEmbedField{
		Name:   isAlive,
		Inline: true,
	}

	embd := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Inventory %s", discord.EmojiInventory),
		Fields: []*discordgo.MessageEmbedField{
			roleField,
			alignmentField,
			coinField,
			abilitiesField,
			anyAbilitiesField,
			perksField,
			itemsField,
			statusesField,
			immunitiesField,
			effectsField,
			deadField,
		},
		Color: discord.ColorThemeDiamond,
	}

	humanReqTime := util.GetEstTimeStamp()
	embd.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Last updated: %s", humanReqTime),
	}

	if host {

		embd.Fields = append(embd.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Luck", discord.EmojiLuck),
			Value:  fmt.Sprintf("%d", inv.Luck),
			Inline: true,
		})

		noteListString := ""
		for i, note := range inv.Notes {
			noteListString += fmt.Sprintf("%d. %s\n", i+1, note)
		}

		embd.Fields = append(embd.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Notes", discord.EmojiNote),
			Value:  noteListString,
			Inline: false,
		})

		embd.Color = discord.ColorThemeAmethyst

	}

	return embd
}
