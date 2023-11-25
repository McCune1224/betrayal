package alliance

import (
	"database/sql"
	"errors"

	"github.com/mccune1224/betrayal/internal/data"
)

func (ah *AllianceHandler) InvitePlayer(ownerID string, inviteeID string, allianceName string) error {
	//
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
		AllianceName: alliance.Name,
		InviterID:    ownerID,
		InviteeID:    inviteeID,
	}

	err = ah.m.Alliances.InsertInvite(invite)
	if err != nil {
		return err
	}
	return nil
}

func (ah *AllianceHandler) AcceptInvite(inviteeID, allianceName string, bypassMemberLimit bool, bypassAllianceLimit bool) error {
	// Check to make sure player isn't already a member of an alliance (unless they have a bypass because of a role perk)
	if !bypassAllianceLimit {
		alliances, err := ah.m.Alliances.GetAllByMemberID(inviteeID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if len(alliances) != 1 {
			return ErrAlreadyAllianceMember
		}
	}

	// Check to make sure invite exists
	invite, err := ah.m.Alliances.GetInviteByInviteeIDAndAllianceName(inviteeID, allianceName)
	if err != nil {
		return err
	}
	if invite.AllianceName == "" {
		return ErrInviteNotFound
	}
	// Get Alliance
	alliance, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil {
		return err
	}

	// Check to make sure alliance has room (max 4 members (3 members + owner))
	if len(alliance.MemberIDs) > 3 && !bypassMemberLimit {
		return ErrAllianceMemberLimitExceeded
	}

	// Delete the invite
	err = ah.m.Alliances.DeleteInvite(invite)
	if err != nil {
		return err
	}
	// Add the player to the alliance
	alliance.MemberIDs = append(alliance.MemberIDs, inviteeID)
	err = ah.m.Alliances.InsertMember(alliance)
	if err != nil {
		return err
	}
	return nil
}
