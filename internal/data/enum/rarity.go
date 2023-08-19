package enum

type RarityEnum string

var Rarity = struct {
	COMMON    RarityEnum
	UNCOMMON  RarityEnum
	RARE      RarityEnum
	EPIC      RarityEnum
	LEGENDARY RarityEnum
	MYTHICAL  RarityEnum
	UNIQUE    RarityEnum
	ULTIMATE  RarityEnum
}{
	COMMON:    "COMMON",
	UNCOMMON:  "UNCOMMON",
	RARE:      "RARE",
	EPIC:      "EPIC",
	LEGENDARY: "LEGENDARY",
	MYTHICAL:  "MYTHICAL",
	UNIQUE:    "UNIQUE",
	ULTIMATE:  "ULTIMATE",
}
