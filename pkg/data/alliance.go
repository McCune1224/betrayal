package data

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Alliance struct {
	ID        int            `db:"id" json:"id"`
	Name      string         `db:"name" json:"name"`
	ChannelID string         `db:"channel_id" json:"channel_id"`
	MemberIDs pq.StringArray `db:"member_ids" json:"member_ids"`
}

type AllianceRequest struct {
	ID          int    `db:"id" json:"id"`
	RequesterID string `db:"requester_id" json:"requester_id"`
	Name        string `db:"name" json:"name"`
}

type AllianceInvite struct {
	ID              int    `db:"id" json:"id"`
	InviterID       string `db:"inviter_id" json:"inviter_id"`
	InviteeID       string `db:"invitee_id" json:"invitee_id"`
	AllianceName    string `db:"alliance_name" json:"alliance_name"`
	InviteeAccepted bool   `db:"invitee_accepted" json:"invitee_accepted"`
}

type AllianceModel struct {
	DB *sqlx.DB
}

func (am *AllianceModel) InsertRequest(req *AllianceRequest) error {
	query := `INSERT INTO alliance_requests ` + PSQLGeneratedInsert(req)
	_, err := am.DB.NamedExec(query, &req)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) GetRequestByName(name string) (*AllianceRequest, error) {
	var req AllianceRequest
	// Case insensitive search (ILIKE)
	query := `SELECT * FROM alliance_requests WHERE name ILIKE $1`
	err := am.DB.Get(&req, query, name)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (am *AllianceModel) GetRequestByRequesterID(name string) (*AllianceRequest, error) {
	var req AllianceRequest
	query := `SELECT * FROM alliance_requests WHERE requester_id=$1`
	err := am.DB.Get(&req, query, name)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (am *AllianceModel) GetAllRequests() ([]AllianceRequest, error) {
	var reqs []AllianceRequest
	query := `SELECT * FROM alliance_requests`
	err := am.DB.Select(&reqs, query)
	if err != nil {
		return nil, err
	}
	return reqs, nil
}

func (ah *AllianceModel) DeleteRequestByName(name string) error {
	query := `DELETE FROM alliance_requests WHERE name=$1`
	_, err := ah.DB.Exec(query, name)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) DeleteRequest(req *AllianceRequest) error {
	query := `DELETE FROM alliance_requests WHERE id=$1`
	_, err := am.DB.Exec(query, req.ID)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) GetAlliances() ([]Alliance, error) {
	var alliances []Alliance
	query := `SELECT * FROM alliances`
	err := am.DB.Select(&alliances, query)
	if err != nil {
		return nil, err
	}
	return alliances, nil
}

func (am *AllianceModel) GetByName(name string) (*Alliance, error) {
	var alliance Alliance
	query := `SELECT * FROM alliances WHERE name ILIKE $1`
	err := am.DB.Get(&alliance, query, name)
	if err != nil {
		return nil, err
	}
	return &alliance, nil
}

func (am *AllianceModel) GetByChannelID(channelID string) (*Alliance, error) {
	var alliance Alliance
	query := `SELECT * FROM alliances WHERE channel_id=$1`
	err := am.DB.Get(&alliance, query, channelID)
	if err != nil {
		return nil, err
	}
	return &alliance, nil
}

func (am *AllianceModel) GetByMemberID(discordID string) (*Alliance, error) {
	var alliance Alliance
	// need to check if any entry has the member id in the member_ids array
	// the member_ids is stored in the psql db as a []varchar
	// so we need to use the @> operator to check if the member id is in the array

	foo := pq.StringArray{discordID}
	query := `SELECT * FROM alliances WHERE member_ids @> $1`
	err := am.DB.Get(&alliance, query, foo)
	if err != nil {
		return nil, err
	}
	return &alliance, nil
}

func (am *AllianceModel) GetAllByMemberID(discordID string) ([]Alliance, error) {
	var alliances []Alliance
	// need to check if any entry has the member id in the member_ids array
	// the member_ids is stored in the psql db as a []varchar
	// so we need to use the @> operator to check if the member id is in the array
	foo := pq.StringArray{discordID}
	query := `SELECT * FROM alliances WHERE member_ids @> $1`
	err := am.DB.Select(&alliances, query, foo)
	if err != nil {
		return nil, err
	}
	return alliances, nil
}

func (am *AllianceModel) Insert(alliance *Alliance) error {
	query := `INSERT INTO alliances ` + PSQLGeneratedInsert(alliance)
	_, err := am.DB.NamedExec(query, &alliance)
	if err != nil {
		return err
	}
	return nil
}

// Delete any associated invites and requests with the alliance
func (am *AllianceModel) Delete(alliance *Alliance) error {
	tx := am.DB.MustBegin()
	// Delete all invites
	inviteQuery := `DELETE FROM alliance_invites WHERE alliance_name=$1`
	_, err := tx.Exec(inviteQuery, alliance.Name)
	if err != nil {
		return err
	}
	requestQuery := `DELETE FROM alliance_requests WHERE name=$1`
	_, err = tx.Exec(requestQuery, alliance.Name)
	if err != nil {
		return err
	}

	// Delete the alliance
	allianceQuery := `DELETE FROM alliances WHERE name=$1`
	_, err = tx.Exec(allianceQuery, alliance.Name)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (am *AllianceModel) InsertMember(alliance *Alliance) error {
	query := `UPDATE alliances SET member_ids=$1  WHERE id=$2`
	_, err := am.DB.Exec(query, alliance.MemberIDs, alliance.ID)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) InsertInvite(invite *AllianceInvite) error {
	query := `INSERT INTO alliance_invites ` + PSQLGeneratedInsert(invite)
	_, err := am.DB.NamedExec(query, &invite)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) UpdateInviteInviteeAccepted(invite *AllianceInvite) error {
	query := `UPDATE alliance_invites SET invitee_accepted=$1 WHERE id=$2`
	_, err := am.DB.Exec(query, invite.InviteeAccepted, invite.ID)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) UpdateAllianceMembers(alliance *Alliance) error {
	if len(alliance.MemberIDs) == 0 {
		// No members left in the alliance, reinitialize the array and insert
		alliance.MemberIDs = pq.StringArray{}
		return am.InsertMember(alliance)
	}

	query := `UPDATE alliances SET member_ids=$1 WHERE id=$1`
	_, err := am.DB.Exec(query, alliance.MemberIDs, alliance.ID)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) GetInviteByInviteeIDAndInviterID(inviteeID, inviterID string) (*AllianceInvite, error) {
	var invite AllianceInvite
	query := `SELECT * FROM alliance_invites WHERE invitee_id=$1 AND inviter_id=$2`
	err := am.DB.Get(&invite, query, inviteeID, inviterID)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (am *AllianceModel) GetInviteByInviteeIDAndAllianceName(inviteeID, allianceName string) (*AllianceInvite, error) {
	var invite AllianceInvite
	query := `SELECT * FROM alliance_invites WHERE invitee_id=$1 AND alliance_name=$2`
	err := am.DB.Get(&invite, query, inviteeID, allianceName)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (am *AllianceModel) GetInvitesByAllianceName(allianceName string) ([]AllianceInvite, error) {
	var invites []AllianceInvite
	query := `SELECT * FROM alliance_invites WHERE alliance_name=$1`
	err := am.DB.Select(&invites, query, allianceName)
	if err != nil {
		return nil, err
	}
	return invites, nil
}

func (am *AllianceModel) GetAllInvitesForUser(userID string) ([]AllianceInvite, error) {
	var invites []AllianceInvite
	query := `SELECT * FROM alliance_invites WHERE invitee_id=$1`
	err := am.DB.Select(&invites, query, userID)
	if err != nil {
		return nil, err
	}
	return invites, nil
}

func (am *AllianceModel) DeleteInvite(invite *AllianceInvite) error {
	query := `DELETE FROM alliance_invites WHERE id=$1`
	_, err := am.DB.Exec(query, invite.ID)
	if err != nil {
		return err
	}
	return nil
}

func (am AllianceModel) DeleteInviteByInviteeIDAndInviterID(inviteeID, inviterID string) error {
	query := `DELETE FROM alliance_invites WHERE invitee_id=$1 AND inviter_id=$2`
	_, err := am.DB.Exec(query, inviteeID, inviterID)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) GetAllInvites() ([]AllianceInvite, error) {
	var invites []AllianceInvite
	query := `SELECT * FROM alliance_invites`
	err := am.DB.Select(&invites, query)
	if err != nil {
		return nil, err
	}
	return invites, nil
}
