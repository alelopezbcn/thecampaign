package board

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHasArcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f := &field{
		cards: []cards.Warrior{
			cards.NewArcher("a1"),
		},
	}
	assert.True(t, f.HasArcher())

	f.cards = []cards.Warrior{cards.NewDragon("d1")}
	assert.False(t, f.HasArcher())
}

func TestHasDragon(t *testing.T) {
	f := &field{
		cards: []cards.Warrior{
			cards.NewDragon("d1"),
		},
	}
	assert.True(t, f.HasDragon())

	f.cards = []cards.Warrior{cards.NewArcher("a1")}
	assert.False(t, f.HasDragon())
}

func TestHasKnight(t *testing.T) {
	f := &field{
		cards: []cards.Warrior{
			cards.NewKnight("d1"),
		},
	}
	assert.True(t, f.HasKnight())

	f.cards = []cards.Warrior{cards.NewArcher("a1")}
	assert.False(t, f.HasKnight())
}

func TestHasMage(t *testing.T) {
	f := &field{
		cards: []cards.Warrior{
			cards.NewMage("d1"),
		},
	}
	assert.True(t, f.HasMage())

	f.cards = []cards.Warrior{cards.NewArcher("a1")}
	assert.False(t, f.HasMage())
}
