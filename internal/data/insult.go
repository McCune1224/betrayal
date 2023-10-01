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

func (im *InsultModel) GetRandom() (*Insult, error) {
	var i Insult

	err := im.DB.Get(
		&i,
		"SELECT * FROM insults ORDER BY RANDOM() LIMIT 1",
	)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (im *InsultModel) Insert(i *Insult) error {

	query := "INSERT INTO insults " + PSQLGeneratedInsert(i)

	_, err := im.DB.NamedExec(query, &i)
	if err != nil {
		return err
	}

	return nil
}

func (im *InsultModel) DeleteInsult(i *Insult) error {
	var err error

	_, err = im.DB.Exec(
		"DELETE FROM insults WHERE insult = $1",
		i.Insult,
	)
	if err != nil {
		return err
	}

	return nil
}
