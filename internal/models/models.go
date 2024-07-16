// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type Alignment string

const (
	AlignmentGOOD    Alignment = "GOOD"
	AlignmentNEUTRAL Alignment = "NEUTRAL"
	AlignmentEVIL    Alignment = "EVIL"
)

func (e *Alignment) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Alignment(s)
	case string:
		*e = Alignment(s)
	default:
		return fmt.Errorf("unsupported scan type for Alignment: %T", src)
	}
	return nil
}

type NullAlignment struct {
	Alignment Alignment `json:"alignment"`
	Valid     bool      `json:"valid"` // Valid is true if Alignment is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAlignment) Scan(value interface{}) error {
	if value == nil {
		ns.Alignment, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Alignment.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAlignment) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Alignment), nil
}

type Rarity string

const (
	RarityCOMMON       Rarity = "COMMON"
	RarityUNCOMMON     Rarity = "UNCOMMON"
	RarityRARE         Rarity = "RARE"
	RarityEPIC         Rarity = "EPIC"
	RarityLEGENDARY    Rarity = "LEGENDARY"
	RarityMYTHICAL     Rarity = "MYTHICAL"
	RarityROLESPECIFIC Rarity = "ROLE_SPECIFIC"
	RarityUNIQUE       Rarity = "UNIQUE"
)

func (e *Rarity) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Rarity(s)
	case string:
		*e = Rarity(s)
	default:
		return fmt.Errorf("unsupported scan type for Rarity: %T", src)
	}
	return nil
}

type NullRarity struct {
	Rarity Rarity `json:"rarity"`
	Valid  bool   `json:"valid"` // Valid is true if Rarity is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRarity) Scan(value interface{}) error {
	if value == nil {
		ns.Rarity, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Rarity.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRarity) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Rarity), nil
}

type AbilityCategory struct {
	AbilityID  int32 `json:"ability_id"`
	CategoryID int32 `json:"category_id"`
}

type AbilityInfo struct {
	ID             int32       `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	DefaultCharges int32       `json:"default_charges"`
	AnyAbility     bool        `json:"any_ability"`
	RoleSpecificID pgtype.Int4 `json:"role_specific_id"`
	Rarity         Rarity      `json:"rarity"`
}

type AdminChannel struct {
	ChannelID string `json:"channel_id"`
}

type Category struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type Item struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rarity      Rarity `json:"rarity"`
	Cost        int32  `json:"cost"`
}

type ItemCategory struct {
	ItemID     int32 `json:"item_id"`
	CategoryID int32 `json:"category_id"`
}

type PerkInfo struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Player struct {
	ID        int64          `json:"id"`
	RoleID    pgtype.Int4    `json:"role_id"`
	Alive     bool           `json:"alive"`
	Coins     int32          `json:"coins"`
	CoinBonus pgtype.Numeric `json:"coin_bonus"`
	Luck      int32          `json:"luck"`
	ItemLimit int32          `json:"item_limit"`
	Alignment Alignment      `json:"alignment"`
}

type PlayerAbility struct {
	PlayerID  int64 `json:"player_id"`
	AbilityID int32 `json:"ability_id"`
	Quantity  int32 `json:"quantity"`
}

type PlayerConfessional struct {
	PlayerID     int64 `json:"player_id"`
	ChannelID    int64 `json:"channel_id"`
	PinMessageID int64 `json:"pin_message_id"`
}

type PlayerImmunity struct {
	PlayerID int64 `json:"player_id"`
	StatusID int32 `json:"status_id"`
}

type PlayerItem struct {
	PlayerID int64 `json:"player_id"`
	ItemID   int32 `json:"item_id"`
	Quantity int32 `json:"quantity"`
}

type PlayerPerk struct {
	PlayerID int64 `json:"player_id"`
	PerkID   int32 `json:"perk_id"`
}

type PlayerStatus struct {
	PlayerID int64 `json:"player_id"`
	StatusID int32 `json:"status_id"`
	Quantity int32 `json:"quantity"`
}

type Role struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Alignment   Alignment `json:"alignment"`
}

type RoleAbility struct {
	RoleID    int32 `json:"role_id"`
	AbilityID int32 `json:"ability_id"`
}

type RolePerk struct {
	RoleID int32 `json:"role_id"`
	PerkID int32 `json:"perk_id"`
}

type Status struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
