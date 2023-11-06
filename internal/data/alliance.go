package data

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Alliance struct {
	ID        int            `db:"id"`
	Name      string         `db:"name"`
	OwnerID   string         `db:"owner_id"`
	ChannelID string         `db:"channel_id"`
	MemberIDs pq.StringArray `db:"member_ids"`
}

type AllianceRequest struct {
	ID          int    `db:"id"`
	RequesterID string `db:"requester_id"`
	Name        string `db:"name"`
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

func (am *AllianceModel) GetRequestByOwnerID(name string) (*AllianceRequest, error) {
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

func (am *AllianceModel) DeleteRequest(req *AllianceRequest) error {
	query := `DELETE FROM alliance_requests WHERE id=$1`
	_, err := am.DB.Exec(query, req.ID)
	if err != nil {
		return err
	}
	return nil
}

func (am *AllianceModel) DeleteRequestByName(name string) error {
	query := `DELETE FROM alliance_requests WHERE name ILIKE $1`
	_, err := am.DB.Exec(query, name)
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

func (am *AllianceModel) Insert(alliance *Alliance) error {
	query := `INSERT INTO alliances ` + PSQLGeneratedInsert(alliance)
	_, err := am.DB.NamedExec(query, &alliance)
	if err != nil {
		return err
	}
	return nil
}
