package data

import "gorm.io/gorm"

// model Status {
//     id             Int       @id @default(autoincrement())
//     name           String    @unique
//     effect         String
//     detailedEffect String?
//     updatedAt      DateTime? @updatedAt
// }

type Status struct {
	gorm.Model
	Name           string `gorm:"unique;not null"`
	Effect         string `gorm:"not null"`
	DetailedEffect string
}

type StatusModel struct {
	DB *gorm.DB
}

//
// model StatusLink {
//     id        Int  @id @default(autoincrement())
//     itemId    Int?
//     abilityId Int?
//     perkId    Int?
//
//     linkType LinkType
//     statuses StatusName[]
//     ability  Ability?     @relation(fields: [abilityId], references: [id])
//     perk     Perk?        @relation(fields: [perkId], references: [id])
//     item     Item?        @relation(fields: [itemId], references: [id])
// }
