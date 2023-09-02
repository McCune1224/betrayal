package data

import (
	"github.com/jmoiron/sqlx"
)

type Insult struct {
	Id        int64  `db:"id"`
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
		"SELECT * FROM insults ORDER BY RANDOM() LIMIT 1",
	)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (m *InsultModel) Insert(i *Insult) error {
	query := `INSERT INTO insults (insult, author_id) VALUES (:insult, :author_id)`

	_, err := m.DB.NamedExec(query, &i)
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
