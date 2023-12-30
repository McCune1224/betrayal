package alliance

import (
	"database/sql"
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/mccune1224/betrayal/internal/discord"
	"github.com/mccune1224/betrayal/pkg/data"
)

var allianceMemberLimit = 4

func (ah *AllianceHandler) InvitePlayer(memberID string, inviteeID string, allianceName string) error {
	// Check to make sure alliance exists
	alliance, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAllianceNotFound
		}
		return err
	}

	// Check to make sure player isn't already invited to an alliance
	_, err = ah.m.Alliances.GetInviteByInviteeIDAndAllianceName(inviteeID, allianceName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Create the invite
	invite := &data.AllianceInvite{
		AllianceName:    alliance.Name,
		InviterID:       memberID,
		InviteeID:       inviteeID,
		InviteeAccepted: false,
	}

	return ah.m.Alliances.InsertInvite(invite)
}

func (ah *AllianceHandler) AcceptInvite(s *discordgo.Session, inviteeID, allianceName string) error {
	// Get Alliance
	alliance, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAllianceNotFound
		}
		return err
	}

	// Get invite
	pendingInvite, err := ah.m.Alliances.GetInviteByInviteeIDAndAllianceName(inviteeID, alliance.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInviteNotFound
		}
		return err
	}

	// Update the invite
	pendingInvite.InviteeAccepted = true
	return ah.m.Alliances.UpdateInviteInviteeAccepted(pendingInvite)
}

func (ah *AllianceHandler) AdminApproveInvite(s *discordgo.Session, allianceName string, inviteeID string) error {
	// Check to make sure alliance exists
	alliance, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAllianceNotFound
		}
		return err
	}
	pendingInvite, err := ah.m.Alliances.GetInviteByInviteeIDAndAllianceName(inviteeID, alliance.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInviteNotFound
		}
		return err
	}

	if !pendingInvite.InviteeAccepted {
		return ErrInviteNotAccepted
	}

	// Add the player to the alliance
	alliance.MemberIDs = append(alliance.MemberIDs, inviteeID)
	err = ah.m.Alliances.InsertMember(alliance)
	if err != nil {
		return err
	}
	err = discord.AddMemberToChannel(s, allianceName, inviteeID)
	if err != nil {
		return err
	}
	return ah.m.Alliances.DeleteInvite(pendingInvite)
}
