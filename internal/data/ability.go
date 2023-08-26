package data

import (
	"github.com/jmoiron/sqlx"
)

const (
	// Unlimited charges on certain abilities
	unlimited = -1
)

// model Ability {
//     id             Int              @id @default(autoincrement())
//     name           String           @unique
//     actionType     ActionType?
//     categories     ActionCategory[]
//     charges        Int // -1 is unlimited
//     isAnyAbility   Boolean          @default(false)
//     isRoleSpecific Boolean?
//     rarity         Rarity?
//     effect         String
//     detailedEffect String?
//     orderPriority  Int              @default(0)
//     showCategories Boolean          @default(true)
//
//     changes            AbilityChange[]
//     abilityAttachments AbilityAttachment?
//
//     updatedAt   DateTime?    @updatedAt
//     statusLinks StatusLink[]
// }
//

// Game representation of an AbilityModel.
type Ability struct {
	Name       string
	ActionType string
	Categories []Category
	// -1 is unlimited
	Charges        int
	IsAnyAbility   bool
	IsRoleSpecific bool
	Rarity         string
	Effect         string
	DetailedEffect string
	OrderPriority  int
	ShowCategories bool
}

type AbilityModel struct {
	DB *sqlx.DB
}

// made since SQL doesn't support string arrays
type Category struct {
	Name string
}

type CategoryModel struct {
	DB *sqlx.DB
}
