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
	Name       string `gorm:"unique;not null"`
	ActionType string `gorm:"not null"`
	Categories string `gorm:"many2many:ability_categories;"`
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

type Category struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
}

type AbilityChange struct {
	gorm.Model
	Ability    Ability `gorm:"foreignKey:AbilityID"`
	AbilityID  uint    `gorm:"not null"`
	Change     string  `gorm:"not null"`
	ChangeType string  `gorm:"not null"`
}

type AbilityAttachment struct {
	gorm.Model
	Ability   Ability `gorm:"foreignKey:AbilityID"`
	AbilityID uint    `gorm:"not null"`
	Roles     []Role  `gorm:"many2many:ability_attachment_roles;"`
}

type AbilityModel struct {
	DB *gorm.DB
}
