package data

import (
	"github.com/jmoiron/sqlx"
)

type Insult struct {
	Insult    string `db:"insult"`
	AuthorID  string `db:"author_id"`
	CreatedAt string `db:"created_at"`
}

type InsultModel struct {
	DB *sqlx.DB
}

func (m *InsultModel) GetRandom() (*Insult, error) {
	var i Insult

	err := m.DB.Get(
		&i,
		"SELECT insult, author_id, created_at FROM insults ORDER BY RANDOM() LIMIT 1",
	)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (m *InsultModel) Insert(i *Insult) error {
	var err error

	_, err = m.DB.Exec(
		"INSERT INTO insults (insult, author_id) VALUES ($1, $2)",
		i.Insult,
		i.AuthorID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *InsultModel) DeleteInsult(i *Insult) error {
	var err error

	_, err = m.DB.Exec(
		"DELETE FROM insults WHERE insult = $1",
		i.Insult,
	)
	if err != nil {
		return err
	}

	return nil
}
