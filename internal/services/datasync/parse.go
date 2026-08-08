package datasync

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mccune1224/betrayal/internal/models"
)

// ParseRolesCSV parses a roles sheet (good/evil/neutral CSV export) into role
// documents. The sheet format is chunked: a role header row (name in col 2,
// description in col 3), then one row per ability, a "Passives:" marker row,
// then one row per passive. Chunks are separated by rows with an empty second
// column. The first chunk is the sheet header and is dropped.
//
// Parse errors are surfaced as warnings with the offending document skipped —
// one malformed row must not take down the whole preview.
func ParseRolesCSV(r io.Reader, alignment models.Alignment) ([]RoleDoc, []string, error) {
	chunks, err := readRoleChunks(r)
	if err != nil {
		return nil, nil, err
	}
	if len(chunks) < 1 {
		return nil, nil, errors.New("no role records found in CSV")
	}

	docs := make([]RoleDoc, 0, len(chunks))
	var warnings []string
	for i, chunk := range chunks {
		doc, warn, err := parseRoleChunk(chunk, alignment)
		warnings = append(warnings, warn...)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("role chunk %d: %v", i+1, err))
			continue
		}
		docs = append(docs, doc)
	}
	return docs, warnings, nil
}

// readRoleChunks splits the CSV into role chunks on rows with an empty second
// column. Empty chunks ARE appended (faithful to the original tool): sheets
// open with a row like "so,,,,,," that yields an empty first chunk, which the
// caller drops via chunks[1:]. Without this, the first role would be lost.
func readRoleChunks(r io.Reader) ([][][]string, error) {
	reader := csv.NewReader(r)
	var chunks [][][]string
	var curr [][]string
	flush := func() {
		chunks = append(chunks, curr)
		curr = nil
	}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			flush()
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < 2 || record[1] == "" {
			flush()
		} else {
			curr = append(curr, record)
		}
	}
	if len(chunks) > 0 {
		chunks = chunks[1:] // leading empty chunk (or sheet header)
	}
	return chunks, nil
}

// parseRoleChunk turns one chunk (role header + abilities + passives) into a
// RoleDoc. Mirrors the archived cmd/data-entry parser, with bounds checks and
// warning accumulation instead of panics.
func parseRoleChunk(chunk [][]string, alignment models.Alignment) (RoleDoc, []string, error) {
	var doc RoleDoc
	var warnings []string

	if len(chunk) < 2 || len(chunk[1]) < 3 {
		return doc, nil, errors.New("role header row too short")
	}
	doc.Name = chunk[1][1]
	doc.Description = chunk[1][2]
	doc.Alignment = alignment

	// Abilities run from row 3 until the "Passives:" marker row.
	idx := 3
	for idx < len(chunk) {
		row := chunk[idx]
		if len(row) > 1 && row[1] == "Passives:" {
			break
		}
		ab, warn, err := parseAbility(row)
		warnings = append(warnings, warn...)
		if err != nil {
			return doc, warnings, fmt.Errorf("ability row %d: %w", idx+1, err)
		}
		doc.Abilities = append(doc.Abilities, ab)
		idx++
	}

	// Perks run after the "Passives:" marker.
	for i := idx + 1; i < len(chunk); i++ {
		row := chunk[i]
		if len(row) < 3 {
			continue
		}
		doc.Perks = append(doc.Perks, PerkDoc{Name: row[1], Description: row[2]})
	}
	return doc, warnings, nil
}

// parseAbility parses one ability row. Column layout (1-indexed as in the
// sheet): 2=name, 3=charges ("∞" allowed), 4=type marker (* / ^ / empty),
// 5=description, 6=categories (slash-separated), 7=rarity (only for * type).
func parseAbility(row []string) (AbilityDoc, []string, error) {
	var doc AbilityDoc
	if len(row) < 7 {
		return doc, nil, fmt.Errorf("expected >=7 columns, got %d", len(row))
	}
	doc.Name = row[1]
	doc.Description = row[4]

	// Charges: "∞" → 999999, otherwise an integer.
	doc.DefaultCharges = infiniteCharges
	if strings.TrimSpace(row[2]) != "∞" {
		charges, err := strconv.Atoi(row[2])
		if err != nil {
			return doc, nil, fmt.Errorf("invalid charge count %q", row[2])
		}
		doc.DefaultCharges = int32(charges)
	}

	// Type marker + rarity.
	switch row[3] {
	case "*":
		doc.AnyAbility = true
		rarity, err := parseRarity(row[6])
		if err != nil {
			return doc, nil, err
		}
		doc.Rarity = rarity
	case "^":
		doc.AnyAbility = true
		doc.Rarity = models.RarityROLESPECIFIC
	case "":
		doc.AnyAbility = false
		doc.Rarity = models.RarityROLESPECIFIC
	default:
		doc.AnyAbility = false
		doc.Rarity = models.RarityROLESPECIFIC
		return doc, []string{fmt.Sprintf("ability %q: unknown type marker %q, defaulting to ROLE_SPECIFIC", doc.Name, row[3])}, nil
	}

	doc.Categories = splitCategories(row[5])
	return doc, nil, nil
}

// ParseItemsCSV parses the items sheet. The first two rows are headers and are
// dropped. Column layout (0-indexed): 1=rarity, 2=name, 3=cost ("X" = free),
// 4=categories (slash-separated), 5=description.
func ParseItemsCSV(r io.Reader) ([]ItemDoc, []string, error) {
	reader := csv.NewReader(r)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	docs := make([]ItemDoc, 0, len(rows))
	var warnings []string
	for i, entry := range rows {
		if i == 0 || i == 1 || len(entry) < 6 {
			if i == 0 || i == 1 {
				continue
			}
			warnings = append(warnings, fmt.Sprintf("item row %d: expected >=6 columns, got %d", i+1, len(entry)))
			continue
		}

		rarity, err := parseRarity(entry[1])
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("item row %d: %v — skipping", i+1, err))
			continue
		}

		cost := int32(0)
		if strings.TrimSpace(entry[3]) != "X" {
			c, err := strconv.ParseInt(entry[3], 10, 32)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("item row %d: invalid cost %q — skipping", i+1, entry[3]))
				continue
			}
			cost = int32(c)
		}

		docs = append(docs, ItemDoc{
			Name:        entry[2],
			Description: entry[5],
			Rarity:      rarity,
			Cost:        cost,
			Categories:  splitCategories(entry[4]),
		})
	}
	return docs, warnings, nil
}

// parseRarity maps a sheet rarity string to the enum, case-insensitively.
func parseRarity(s string) (models.Rarity, error) {
	switch models.Rarity(strings.ToUpper(strings.TrimSpace(s))) {
	case models.RarityCOMMON,
		models.RarityUNCOMMON,
		models.RarityRARE,
		models.RarityEPIC,
		models.RarityLEGENDARY,
		models.RarityMYTHICAL,
		models.RarityUNIQUE:
		return models.Rarity(strings.ToUpper(strings.TrimSpace(s))), nil
	default:
		return "", fmt.Errorf("unknown rarity %q", s)
	}
}

// splitCategories splits and trims a slash-separated category list.
func splitCategories(s string) []string {
	parts := strings.Split(s, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
