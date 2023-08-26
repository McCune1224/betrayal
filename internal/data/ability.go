package data

import (
	"gorm.io/gorm"
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
	gorm.Model
	Name       string     `gorm:"unique;not null"`
	ActionType string     `gorm:"not null"`
	Categories []Category `gorm:"many2many:ability_categories;"`
	// -1 is unlimited
	Charges        int    `gorm:"not null"`
	IsAnyAbility   bool   `gorm:"not null;default:false"`
	IsRoleSpecific bool   `gorm:"not null"`
	Rarity         string `gorm:"not null"`
	Effect         string `gorm:"not null"`
	DetailedEffect string
	OrderPriority  int  `gorm:"not null;default:0"`
	ShowCategories bool `gorm:"not null;default:true"`
}

type AbilityModel struct {
	DB *gorm.DB
}

// made since SQL doesn't support string arrays
type Category struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
}

type CategoryModel struct {
	DB *gorm.DB
}
