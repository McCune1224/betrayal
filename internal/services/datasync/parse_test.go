package datasync_test

import (
	"strings"
	"testing"

	"github.com/mccune1224/betrayal/internal/models"
	"github.com/mccune1224/betrayal/internal/services/datasync"
	"github.com/stretchr/testify/require"
)

// roleCSV mirrors the Google Sheets role export format: a leading header
// chunk (dropped), then chunked roles split on empty second columns. Within a
// chunk: role name/desc in row 1, abilities from row 3, "Passives:" marker,
// then passive rows.
// roleCSV mirrors the real Google Sheets role export (verified against the
// live GOOD_ROLES sheet, 2026-08): a leading "so,,,,,," row that yields the
// dropped empty chunk, then per role: a "Name/Description" label row, the role
// row, an "Abilities:" marker row, ability rows, a "Passives:" marker row,
// then passive rows. Chunks are separated by blank rows.
const roleCSV = `so,,,,,,
,Name ,Description,,,,
,RoleA,Role description A,,,,
,Abilities:,Charges,Type,Description,Categories,Rarity (if AA)
,Ability One,3,*,Does a thing,Stealth/Combat,COMMON
,Ability Two,∞,^,Another thing,Social,
,Passives:,Description,,,,
,Passive One,Passive desc one,,,,
,,,,,,
,Name ,Description,,,,
,RoleB,Role description B,,,,
,Abilities:,Charges,Type,Description,Categories,Rarity (if AA)
,Ability Three,1,,Third thing,Magic,RARE
,Passives:,Description,,,,
,Passive Two,Passive desc two,,,,
`

func TestParseRolesCSV(t *testing.T) {
	docs, warnings, err := datasync.ParseRolesCSV(strings.NewReader(roleCSV), models.AlignmentGOOD)
	require.NoError(t, err)
	require.Empty(t, warnings)
	require.Len(t, docs, 2)

	// RoleA with two abilities and one passive.
	a := docs[0]
	require.Equal(t, "RoleA", a.Name)
	require.Equal(t, "Role description A", a.Description)
	require.Equal(t, models.AlignmentGOOD, a.Alignment)
	require.Len(t, a.Abilities, 2)
	require.Len(t, a.Perks, 1)

	one := a.Abilities[0]
	require.Equal(t, "Ability One", one.Name)
	require.Equal(t, "Does a thing", one.Description)
	require.Equal(t, int32(3), one.DefaultCharges)
	require.True(t, one.AnyAbility)
	require.Equal(t, models.RarityCOMMON, one.Rarity)
	require.Equal(t, []string{"Stealth", "Combat"}, one.Categories)

	two := a.Abilities[1]
	require.Equal(t, int32(999999), two.DefaultCharges, "∞ parses to the infinite-charges encoding")
	require.True(t, two.AnyAbility)
	require.Equal(t, models.RarityROLESPECIFIC, two.Rarity)

	require.Equal(t, "Passive One", a.Perks[0].Name)
	require.Equal(t, "Passive desc one", a.Perks[0].Description)

	// RoleB single ability, one passive.
	b := docs[1]
	require.Equal(t, "RoleB", b.Name)
	require.Equal(t, models.AlignmentGOOD, b.Alignment)
	require.Len(t, b.Abilities, 1)
	require.Equal(t, "Ability Three", b.Abilities[0].Name)
	require.False(t, b.Abilities[0].AnyAbility)
	require.Equal(t, models.RarityROLESPECIFIC, b.Abilities[0].Rarity)
	require.Len(t, b.Perks, 1)
	require.Equal(t, "Passive Two", b.Perks[0].Name)
}

func TestParseRolesCSV_NoLeadingSORow(t *testing.T) {
	// Sheets without the leading "so,,,,,," row must not lose their first role.
	csv := `,Name ,Description,,,,
,RoleFirst,First role desc,,,,
,Abilities:,Charges,Type,Description,Categories,Rarity (if AA)
,Passives:,Description,,,,
`
	docs, warnings, err := datasync.ParseRolesCSV(strings.NewReader(csv), models.AlignmentGOOD)
	require.NoError(t, err)
	require.Empty(t, warnings)
	require.Len(t, docs, 1, "first role survives without the so row")
	require.Equal(t, "RoleFirst", docs[0].Name)
}

func TestParseRolesCSV_TrailingBlankRowNoWarning(t *testing.T) {
	csv := `so,,,,,,
,Name ,Description,,,,
,RoleA,Role description A,,,,
,Abilities:,Charges,Type,Description,Categories,Rarity (if AA)
,Passives:,Description,,,,
,,,,,,
`
	docs, warnings, err := datasync.ParseRolesCSV(strings.NewReader(csv), models.AlignmentGOOD)
	require.NoError(t, err)
	require.Empty(t, warnings, "trailing blank separator must not produce a chunk warning")
	require.Len(t, docs, 1)
	require.Equal(t, "RoleA", docs[0].Name)
}

func TestParseRolesCSV_TrimsNames(t *testing.T) {
	csv := `so,,,,,,
,Name ,Description,,,,
,RoleA ,Role description A,,,,
,Abilities:,Charges,Type,Description,Categories,Rarity (if AA)
,Ability One ,3,*,Does a thing,Stealth/Combat,COMMON
,Passives:,Description,,,,
,Passive One ,Passive desc one,,,,
`
	docs, warnings, err := datasync.ParseRolesCSV(strings.NewReader(csv), models.AlignmentGOOD)
	require.NoError(t, err)
	require.Empty(t, warnings)
	require.Len(t, docs, 1)
	require.Equal(t, "RoleA", docs[0].Name, "trailing whitespace trimmed from role name")
	require.Equal(t, "Ability One", docs[0].Abilities[0].Name, "trailing whitespace trimmed from ability name")
	require.Equal(t, "Passive One", docs[0].Perks[0].Name, "trailing whitespace trimmed from perk name")
}

func TestParseRolesCSV_UnknownRarityWarnsAndSkips(t *testing.T) {
	csv := `so,,,,,,
,Name ,Description,,,,
,RoleX,Role desc X,,,,
,Abilities:,Charges,Type,Description,Categories,Rarity (if AA)
,Ability Bad,1,*,Thing,,NOT_A_RARITY
,Passives:,Description,,,,
`
	docs, warnings, err := datasync.ParseRolesCSV(strings.NewReader(csv), models.AlignmentEVIL)
	require.NoError(t, err)
	require.Len(t, docs, 0, "role with an unparseable ability is skipped")
	require.NotEmpty(t, warnings)
	require.Contains(t, warnings[0], "unknown rarity")
}

func TestParseRolesCSV_UnknownTypeMarkerWarns(t *testing.T) {
	csv := `so,,,,,,
,Name ,Description,,,,
,RoleY,Role desc Y,,,,
,Abilities:,Charges,Type,Description,Categories,Rarity (if AA)
,Ability Odd,1,?,Thing,,COMMON
,Passives:,Description,,,,
`
	docs, warnings, err := datasync.ParseRolesCSV(strings.NewReader(csv), models.AlignmentNEUTRAL)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Len(t, docs[0].Abilities, 1)
	require.False(t, docs[0].Abilities[0].AnyAbility)
	require.Equal(t, models.RarityROLESPECIFIC, docs[0].Abilities[0].Rarity)
	require.NotEmpty(t, warnings)
	require.Contains(t, warnings[0], "unknown type marker")
}

// itemCSV mirrors the items sheet: two header rows, then rarity/name/cost/
// categories/description per row.
const itemCSV = `hdr,hdr,hdr,hdr,hdr,hdr
hdr,hdr,hdr,hdr,hdr,hdr
,RARE,Sword of Testing,50,Weapons/Melee,Stabby sword
,COMMON,Free Thing,X,,No cost
,BOGUS,Invalid Item,5,,Bad rarity row is skipped
,COMMON,Costly Thing,notanumber,,Skipped on bad cost
`

func TestParseItemsCSV(t *testing.T) {
	docs, warnings, err := datasync.ParseItemsCSV(strings.NewReader(itemCSV))
	require.NoError(t, err)
	require.Len(t, docs, 2)
	require.Len(t, warnings, 2)

	sword := docs[0]
	require.Equal(t, "Sword of Testing", sword.Name)
	require.Equal(t, models.RarityRARE, sword.Rarity)
	require.Equal(t, int32(50), sword.Cost)
	require.Equal(t, []string{"Weapons", "Melee"}, sword.Categories)
	require.Equal(t, "Stabby sword", sword.Description)

	free := docs[1]
	require.Equal(t, "Free Thing", free.Name)
	require.Equal(t, int32(0), free.Cost, "X cost maps to 0")
	require.Empty(t, free.Categories)

	require.Contains(t, warnings[0], "unknown rarity")
	require.Contains(t, warnings[1], "invalid cost")
}
