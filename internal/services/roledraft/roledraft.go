package roledraft

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mccune1224/betrayal/internal/models"
)

// ActiveRoleNames is the curated role list used for a new game setup.
var ActiveRoleNames = []string{"Agent", "Amalgamation", "Anarchist", "Analyst", "Backstabber", "Arsonist", "Biker", "Bard", "Bartender", "Cerberus", "Bomber", "Consort", "Detective", "Cheater", "Director", "Fisherman", "Entertainer", "Doll", "Gunman", "Empress", "Forsaken Angel", "Hero", "Ghost", "Gatekeeper", "Hydra", "Goliath", "Hacker", "Judge", "Incubus", "Highwayman", "Knight", "Magician", "Hunter", "The Major", "Masochist", "Jester", "Medium", "Mercenary", "Juggernaut", "Nurse", "Mimic", "Overlord", "Seraph", "Pathologist", "Parasite", "Terminal", "Salesman", "Phantom", "Time Traveler", "Siren", "Psychotherapist", "Undercover", "Sidekick", "Slaughterer", "Wizard", "Villager", "Threatener", "Yeti", "Wanderer", "Witchdoctor"}

type Pool struct {
	DeceptionOptions [][]models.Role
	RandomPool       []models.Role
}

func LoadRoles(ctx context.Context, pool *pgxpool.Pool) ([]models.Role, error) {
	return models.New(pool).ListRolesByName(ctx, ActiveRoleNames)
}

func Generate(roles []models.Role, players, deceptionists int) (*Pool, error) {
	if players < 0 || players > len(roles) {
		return nil, fmt.Errorf("player count must be between 0 and %d", len(roles))
	}
	if deceptionists < 0 {
		return nil, fmt.Errorf("deceptionist count cannot be negative")
	}
	good, evil, neutral := Group(roles)
	max := deceptionists
	if max > len(good) {
		max = len(good)
	}
	if max > len(evil) {
		max = len(evil)
	}
	if max > len(neutral) {
		max = len(neutral)
	}
	out := &Pool{}
	gp, ep, np := rand.Perm(len(good)), rand.Perm(len(evil)), rand.Perm(len(neutral))
	for i := 0; i < max; i++ {
		out.DeceptionOptions = append(out.DeceptionOptions, []models.Role{good[gp[i]], neutral[np[i]], evil[ep[i]]})
	}
	rp := rand.Perm(len(roles))
	for i := 0; i < players; i++ {
		out.RandomPool = append(out.RandomPool, roles[rp[i]])
	}
	return out, nil
}

func Group(roles []models.Role) (good, evil, neutral []models.Role) {
	for _, role := range roles {
		switch role.Alignment {
		case models.AlignmentGOOD:
			good = append(good, role)
		case models.AlignmentEVIL:
			evil = append(evil, role)
		case models.AlignmentNEUTRAL:
			neutral = append(neutral, role)
		}
	}
	return
}
