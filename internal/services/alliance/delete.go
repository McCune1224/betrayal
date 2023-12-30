package alliance

import (
	"database/sql"
	"errors"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/pkg/data"
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

func (ah *AllianceHandler) LeaveAlliance(currAlliance *data.Alliance, memberID string, s *discordgo.Session) error {
	memberIndex := slices.Index(currAlliance.MemberIDs, memberID)
	if memberIndex == -1 {
		return ErrMemberNotFound
	}
	currAlliance.MemberIDs = append(currAlliance.MemberIDs[:memberIndex], currAlliance.MemberIDs[memberIndex+1:]...)

	// Remove the member from the channel permissions to see the channel
	err := s.ChannelPermissionDelete(currAlliance.ChannelID, memberID)
	if err != nil {
		return err
	}
	return ah.m.Alliances.UpdateAllianceMembers(currAlliance)
}
