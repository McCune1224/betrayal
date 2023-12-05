package main

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
)

type csvNewRole struct {
	Role         data.Role
	Abilities    []data.Ability
	AnyAbilities []data.AnyAbility
	Perks        []data.Perk
}

func (cnr *csvNewRole) InsertRole(csv [][]string) error {
	return nil
}

func (cnr *csvNewRole) InsertAbilities(csv [][]string) error {
	return nil
}

func (cnr *csvNewRole) InsertAnyAbilities(csv [][]string) error {
	return nil
}

func (cnr *csvNewRole) InsertPerks(csv [][]string) error {
	return nil
}

func (*csvBuilder) BuildNewRoleCSV(csv [][]string, alignment string) ([]csvNewRole, error) {
	if len(csv) == 0 {
		return nil, errors.New("csv is empty")
	}

	roles := []csvNewRole{}

	chunk := [][]string{}
	superChunk := [][][]string{}
	for i := 1; i < len(csv); i++ {
		// edge case for last line
		if i == len(csv)-1 {
			superChunk = append(superChunk, chunk)
			break
		}

		line := csv[i][1:]
		chunk = append(chunk, line)
		if line[0] == "" {
			superChunk = append(superChunk, chunk)
			chunk = [][]string{}
		}
	}

	for _, chunk := range superChunk {
		role := csvNewRole{}
		role.Role.Name = chunk[1][0]
		role.Role.Description = chunk[1][1]
		role.Role.Alignment = alignment

		abIdx := 4
		for chunk[abIdx][0] != "Perks:" {
			chargeStr := chunk[abIdx][1]
			// ∞
			charge := -1
			if chargeStr != "∞" {
				chargeParse, err := strconv.Atoi(chunk[abIdx][1])
				if err != nil {
					log.Fatal(err)
				}
				charge = chargeParse
			}
			categories := pq.StringArray(strings.Split(chunk[abIdx][4], "/"))

			ability := data.Ability{
				Name:        chunk[abIdx][0],
				Charges:     charge,
				Description: chunk[abIdx][3],
				Categories:  categories,
			}
			if chunk[abIdx][2] != "" {
				ability.AnyAbility = true

				aa := data.AnyAbility{
					Name:         chunk[abIdx][0],
					Description:  chunk[abIdx][3],
					Categories:   categories,
					Rarity:       chunk[abIdx][5],
					RoleSpecific: role.Role.Name,
				}
				role.AnyAbilities = append(role.AnyAbilities, aa)
			} else {
				ability.AnyAbility = false
			}

			role.Abilities = append(role.Abilities, ability)
			abIdx++
		}
		perkIdx := abIdx + 1
		for perkIdx < len(chunk) {
			perk := data.Perk{
				Name:        chunk[perkIdx][0],
				Description: chunk[perkIdx][1],
			}
			if perk.Name == "" {
				break
			}
			role.Perks = append(role.Perks, perk)
			perkIdx++
		}
		roles = append(roles, role)
	}

	return roles, nil
}
