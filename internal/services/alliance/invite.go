package alliance

import (
	"github.com/mccune1224/betrayal/internal/data"
)

func (ah *AllianceHandler) InvitePlayer(ownerID string, inviteeID string, allianceName string) error {
	// Check to make sure player is owner of an alliance
	existingAlliance, err := ah.m.Alliances.GetByOwnerID(ownerID)
	if err != nil {
		return err
	}

	if existingAlliance.Name == "" {
		return ErrAllianceNotFound
	}

	// Check to make sure alliance exists
	alliance, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil {
		return err
	}

	if alliance.Name == "" {
		return ErrAllianceNotFound
	}

	// Check to make sure player isn't already a member of an alliance
	existingMember, err := ah.m.Alliances.GetByMemberID(inviteeID)
	if err != nil {
		return err
	}

	if existingMember.Name != "" {
		return ErrMemberAlreadyExists
	}

	// Check to make sure player isn't already invited to an alliance
	existingInvite, err := ah.m.Alliances.GetInviteByInviteeIDAndInviterID(inviteeID, ownerID)
	if err != nil {
		return err
	}

	if existingInvite.AllianceName != "" {
		return ErrInviteAlreadyExists
	}

	// Create the invite
	invite := &data.AllianceInvite{
		AllianceName: allianceName,
		InviterID:    ownerID,
		InviteeID:    inviteeID,
	}

	err = ah.m.Alliances.InsertInvite(invite)
	if err != nil {
		return err
	}

	return nil
}

func (ah *AllianceHandler) AcceptInvite(inviteeID, allianceName string, bypass ...bool) error {
	bypassLimit := false
	if len(bypass) > 0 {
		bypassLimit = bypass[0]
	}
	// Check to make sure player isn't already a member of an alliance
	existingMember, err := ah.m.Alliances.GetByMemberID(inviteeID)
	if err != nil {
		return err
	}
	if existingMember.Name != "" {
		return ErrMemberAlreadyExists
	}
	// Check to make sure invite exists
	invite, err := ah.m.Alliances.GetInviteByInviteeIDAndAllianceName(inviteeID, allianceName)
	if err != nil {
		return err
	}
	if invite.AllianceName == "" {
		return ErrMemberAlreadyExists
	}
	// Get Alliance
	alliance, err := ah.m.Alliances.GetByName(allianceName)
	if err != nil {
		return err
	}

	// Check to make sure alliance has room (max 4 members (3 members + owner))
	if len(alliance.MemberIDs) > 3 && !bypassLimit {
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