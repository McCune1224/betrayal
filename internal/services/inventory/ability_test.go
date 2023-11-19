package inventory

import (
	"testing"

	"github.com/lib/pq"
	"github.com/mccune1224/betrayal/internal/data"
)

func TestInvetoryAbility(t *testing.T) {
	tAB := &data.Ability{
		Name:    "Test Ability",
		Charges: 2,
	}

	tIv := NewInventoryAbility(tAB)

	abList := []pq.StringArray{}

	abList = append(abList, pq.StringArray{tIv.Name()})
}
