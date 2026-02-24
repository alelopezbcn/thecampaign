package board

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

func TestHasWarriorType(t *testing.T) {
	tests := []struct {
		name        string
		fieldCards  []cards.Warrior
		queryType   types.WarriorType
		wantPresent bool
	}{
		{"Archer present", []cards.Warrior{cards.NewArcher("a1")}, types.ArcherWarriorType, true},
		{"Archer absent", []cards.Warrior{cards.NewDragon("d1")}, types.ArcherWarriorType, false},
		{"Dragon present", []cards.Warrior{cards.NewDragon("d1")}, types.DragonWarriorType, true},
		{"Dragon absent", []cards.Warrior{cards.NewArcher("a1")}, types.DragonWarriorType, false},
		{"Knight present", []cards.Warrior{cards.NewKnight("k1")}, types.KnightWarriorType, true},
		{"Knight absent", []cards.Warrior{cards.NewArcher("a1")}, types.KnightWarriorType, false},
		{"Mage present", []cards.Warrior{cards.NewMage("m1")}, types.MageWarriorType, true},
		{"Mage absent", []cards.Warrior{cards.NewArcher("a1")}, types.MageWarriorType, false},
		{"Mercenary present", []cards.Warrior{cards.NewMercenary("mc1")}, types.MercenaryWarriorType, true},
		{"Mercenary absent", []cards.Warrior{cards.NewArcher("a1")}, types.MercenaryWarriorType, false},
		{"Empty field", []cards.Warrior{}, types.KnightWarriorType, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &field{cards: tt.fieldCards}
			assert.Equal(t, tt.wantPresent, f.HasWarriorType(tt.queryType))
		})
	}
}
