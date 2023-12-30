package data

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/util"
)

type Item struct {
	ID          int64          `db:"id" json:"id"`
	Name        string         `db:"name" json:"name"`
	Description string         `db:"description" json:"description"`
	Cost        int64          `db:"cost" json:"cost"`
	Rarity      string         `db:"rarity" json:"rarity"`
	Categories  pq.StringArray `db:"categories" json:"categories"`
	CreatedAt   string         `db:"created_at" json:"created_at"`
}

type ItemModel struct {
	DB *sqlx.DB
}

func (im *ItemModel) Insert(i *Item) (int64, error) {
	query := `INSERT INTO items 
    (name, description, cost, rarity, categories) 
    VALUES ($1, $2, $3, $4, $5) RETURNING id`

	_, err := im.DB.Exec(query, i.Name, i.Description, i.Cost, i.Rarity, i.Categories)
	if err != nil {
		return -1, err
	}

	var lastInsert Item
	err = im.DB.Get(&lastInsert, "SELECT * FROM items ORDER BY id DESC LIMIT 1")

	return lastInsert.ID, nil
}

func (im *ItemModel) Get(id int64) (*Item, error) {
	var item Item
	err := im.DB.Get(&item, "SELECT * FROM items WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (im *ItemModel) GetByName(name string) (*Item, error) {
	var item Item

	err := im.DB.Get(&item, "SELECT * FROM items WHERE name ILIKE $1", name)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (im *ItemModel) GetByFuzzy(name string) (*Item, error) {
	var item Item
	im.DB.Get(&item, "SELECT * FROM items WHERE name ILIKE $1", name)
	if item.Name != "" {
		return &item, nil
	}
	var itemChoices []Item
	if len(name) < 2 {
		return nil, errors.New("serach term must be at least 2 characters")
	}
	err := im.DB.Select(&itemChoices, "SELECT * FROM items")
	if err != nil {
		return nil, err
	}
	strChoices := make([]string, len(itemChoices))
	for i, item := range itemChoices {
		strChoices[i] = item.Name
	}
	best, _ := util.FuzzyFind(name, strChoices)
	for _, i := range itemChoices {
		if i.Name == best {
			item = i
		}
	}
	return &item, nil
}

func (im *ItemModel) GetByRarity(rarity string) ([]Item, error) {
	var items []Item

	err := im.DB.Select(&items, "SELECT * FROM items WHERE rarity ILIKE $1", rarity)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (im *ItemModel) GetAll() ([]Item, error) {
	var items []Item

	err := im.DB.Select(&items, "SELECT * FROM items")
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (im *ItemModel) GetRandomByRarity(rarity string) (*Item, error) {
	var item Item

	err := im.DB.Get(
		&item,
		"SELECT * FROM items WHERE rarity ILIKE $1 ORDER BY RANDOM() LIMIT 1",
		rarity,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (im *ItemModel) Update(item *Item) error {
	query := `UPDATE items SET name = $1, description = $2, cost = $3, categories = $4 WHERE id = $5`
	_, err := im.DB.Exec(query, item.Name, item.Description, item.Cost, item.Categories, item.ID)
	if err != nil {
		return err
	}
	return nil
}

func (im *ItemModel) Delete(id int64) error {
	_, err := im.DB.Exec("DELETE FROM items WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}
