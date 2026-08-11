package api

import (
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
)

type DiscordResourceHandler struct {
	discord *discordgo.Session
}

func NewDiscordResourceHandler(discord *discordgo.Session) *DiscordResourceHandler {
	return &DiscordResourceHandler{discord: discord}
}

type DiscordGuildDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DiscordChannelDTO struct {
	ID       string `json:"id"`
	GuildID  string `json:"guild_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Category string `json:"category,omitempty"`
}

type DiscordMemberDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
	Bot      bool   `json:"bot"`
}

func (h *DiscordResourceHandler) Resources(c echo.Context) error {
	if h.discord == nil {
		WriteError(c.Response(), http.StatusServiceUnavailable, "discord_unavailable", "Discord is disabled or not connected", nil)
		return nil
	}
	guilds := make([]DiscordGuildDTO, 0)
	channels := make([]DiscordChannelDTO, 0)
	members := make([]DiscordMemberDTO, 0)
	if h.discord.State == nil {
		WriteError(c.Response(), http.StatusServiceUnavailable, "discord_unavailable", "Discord state is not ready", nil)
		return nil
	}
	for _, guild := range h.discord.State.Guilds {
		if guild == nil {
			continue
		}
		guilds = append(guilds, DiscordGuildDTO{ID: guild.ID, Name: guild.Name})
		guildChannels, err := h.discord.GuildChannels(guild.ID)
		if err == nil {
			for _, channel := range guildChannels {
				if channel == nil || channel.Type == discordgo.ChannelTypeGuildCategory {
					continue
				}
				category := ""
				for _, candidate := range guildChannels {
					if candidate != nil && candidate.ID == channel.ParentID {
						category = candidate.Name
						break
					}
				}
				channels = append(channels, DiscordChannelDTO{ID: channel.ID, GuildID: guild.ID, Name: channel.Name, Type: strconv.Itoa(int(channel.Type)), Category: category})
			}
		}
		guildMembers, err := h.discord.GuildMembers(guild.ID, "", 1000)
		if err == nil {
			for _, member := range guildMembers {
				if member == nil || member.User == nil {
					continue
				}
				members = append(members, DiscordMemberDTO{ID: member.User.ID, Username: member.User.Username, Nickname: member.Nick, Bot: member.User.Bot})
			}
		}
	}
	WriteJSON(c.Response(), http.StatusOK, map[string]any{"guilds": guilds, "channels": channels, "members": members})
	return nil
}
