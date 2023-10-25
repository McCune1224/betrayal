package data

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Alliance struct {
	ID      int            `db:"id"`
	Name    string         `db:"name"`
	OwnerID string         `db:"owner_id"`
	Members pq.StringArray `db:"members"`
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
	query := `SELECT * FROM alliance_requests WHERE name=$1`
	err := am.DB.Get(&req, query, name)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (am *AbilityModel) DeleteRequest(req *AllianceRequest) error {
	query := `DELETE FROM alliance_requests WHERE id=$1`
	_, err := am.DB.Exec(query, req.ID)
	if err != nil {
		return err
	}
	return nil
}

func (am *AbilityModel) GetAllRequests() ([]AllianceRequest, error) {
	var reqs []AllianceRequest
	query := `SELECT * FROM alliance_requests`
	err := am.DB.Select(&reqs, query)
	if err != nil {
		return nil, err
	}
	return reqs, nil
}
