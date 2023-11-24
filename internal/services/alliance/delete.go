package alliance

import (
	"database/sql"
	"errors"

	"github.com/bwmarrin/discordgo"
)

func (ah *AllianceHandler) DeleteAlliance(allianceName string, s *discordgo.Session) error {
	existing, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAllianceNotFound
		}
		return err
	}
	_, err = s.ChannelDelete(existing.ChannelID)
	if err != nil {
		return err
	}
	return ah.m.Alliances.Delete(existing)
}
