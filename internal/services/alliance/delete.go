package alliance

import (
	"database/sql"
	"errors"
	"slices"

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

func (ah *AllianceHandler) LeaveAlliance(allianceName string, memberID string, s *discordgo.Session) error {
	existing, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAllianceNotFound
		}
		return err
	}
	memberIndex := slices.Index(existing.MemberIDs, memberID)
	if memberIndex == -1 {
		return ErrMemberNotFound
	}
	existing.MemberIDs = append(existing.MemberIDs[:memberIndex], existing.MemberIDs[memberIndex+1:]...)

	// Remove the member from the channel permissions to see the channel
	err = s.ChannelPermissionDelete(existing.ChannelID, memberID)
	if err != nil {
		return err
	}
	return ah.m.Alliances.UpdateAllianceMembers(existing)
}
