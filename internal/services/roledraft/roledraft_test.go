package roledraft

import (
	"github.com/mccune1224/betrayal/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
)

func testRoles() []models.Role {
	return []models.Role{{Name: "Good 1", Alignment: models.AlignmentGOOD}, {Name: "Good 2", Alignment: models.AlignmentGOOD}, {Name: "Evil 1", Alignment: models.AlignmentEVIL}, {Name: "Evil 2", Alignment: models.AlignmentEVIL}, {Name: "Neutral 1", Alignment: models.AlignmentNEUTRAL}, {Name: "Neutral 2", Alignment: models.AlignmentNEUTRAL}}
}
func TestGroup(t *testing.T) {
	g, e, n := Group(testRoles())
	require.Len(t, g, 2)
	require.Len(t, e, 2)
	require.Len(t, n, 2)
}
func TestGenerateCapsDeceptionistsAndSelectsPool(t *testing.T) {
	p, err := Generate(testRoles(), 4, 99)
	require.NoError(t, err)
	require.Len(t, p.RandomPool, 4)
	require.Len(t, p.DeceptionOptions, 2)
	for _, opt := range p.DeceptionOptions {
		require.Len(t, opt, 3)
	}
}
func TestGenerateRejectsTooManyPlayers(t *testing.T) {
	_, err := Generate(testRoles(), 7, 0)
	require.Error(t, err)
}
