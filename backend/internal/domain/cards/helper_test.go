package cards

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// countType counts how many cards in the slice satisfy the type constraint T.
func countType[T any](cs []Card) int {
	n := 0
	for _, c := range cs {
		if _, ok := c.(T); ok {
			n++
		}
	}
	return n
}

// --- WarriorsCards ---

func TestDealer_WarriorsCards_TotalIsWarriorsTimesThree(t *testing.T) {
	for _, w := range []int{0, 3, 5, 7} {
		d := NewDealer(DeckConfig{Warriors: w})
		got := d.WarriorsCards(0)
		assert.Len(t, got, w*3, "warriors=%d", w)
	}
}

func TestDealer_WarriorsCards_EvenSplitAcrossTypes(t *testing.T) {
	d := NewDealer(DeckConfig{Warriors: 4})
	got := d.WarriorsCards(0)

	assert.Equal(t, 4, countType[*knight](got), "knights")
	assert.Equal(t, 4, countType[*archer](got), "archers")
	assert.Equal(t, 4, countType[*mage](got), "mages")
}

// --- OtherCards — special cards ---

func TestDealer_OtherCards_SpecialCardCountsMatchConfig(t *testing.T) {
	cfg := DeckConfig{
		Dragons:           2,
		SpecialPowers:     3,
		Spies:             1,
		Thieves:           2,
		Sabotages:         1,
		Catapults:         2,
		Fortresses:        1,
		Harpoons:          2,
		BloodRains:        1,
		Resurrections:     2,
		Ambushes:          1,
		Desertions:        3,
		ConstructionCards: 1,
	}
	d := NewDealer(cfg)
	other := d.OtherCards(0)

	assert.Equal(t, 2, countType[*dragon](other), "dragons")
	assert.Equal(t, 3, countType[SpecialPower](other), "special powers")
	assert.Equal(t, 1, countType[Spy](other), "spies")
	assert.Equal(t, 2, countType[Thief](other), "thieves")
	assert.Equal(t, 1, countType[Sabotage](other), "sabotages")
	assert.Equal(t, 2, countType[Catapult](other), "catapults")
	assert.Equal(t, 1, countType[Fortress](other), "fortresses")
	assert.Equal(t, 2, countType[Harpoon](other), "harpoons")
	assert.Equal(t, 1, countType[BloodRain](other), "blood rains")
	assert.Equal(t, 2, countType[Resurrection](other), "resurrections")
	assert.Equal(t, 1, countType[Ambush](other), "ambushes")
	assert.Equal(t, 3, countType[Desertion](other), "desertions")
}

// --- OtherCards — construction cards ---

func TestDealer_OtherCards_ConstructionCards_OneCopyOfValueOne(t *testing.T) {
	d := NewDealer(DeckConfig{ConstructionCards: 1})
	other := d.OtherCards(0)

	// 1 value-1 card + 8 fixed cards for values 2-9 = 9 each
	// gold also gets 3 extra copies at values 5, 7, 9 = 12 total
	assert.Equal(t, 12, countType[*gold](other), "gold")
	assert.Equal(t, 9, countType[*sword](other), "swords")
	assert.Equal(t, 9, countType[*arrow](other), "arrows")
	assert.Equal(t, 9, countType[*poison](other), "poisons")
}

func TestDealer_OtherCards_ConstructionCards_TwoCopiesOfValueOne(t *testing.T) {
	d := NewDealer(DeckConfig{ConstructionCards: 2})
	other := d.OtherCards(0)

	// 2 value-1 cards + 8 fixed cards for values 2-9 = 10 each
	// gold also gets 3 extra copies at values 5, 7, 9 = 13 total
	assert.Equal(t, 13, countType[*gold](other), "gold")
	assert.Equal(t, 10, countType[*sword](other), "swords")
	assert.Equal(t, 10, countType[*arrow](other), "arrows")
	assert.Equal(t, 10, countType[*poison](other), "poisons")
}

// --- Desertion edge cases ---

func TestDealer_OtherCards_Desertions_ZeroWhenNotConfigured(t *testing.T) {
	d := NewDealer(DeckConfig{Desertions: 0, ConstructionCards: 1})
	other := d.OtherCards(0)

	assert.Equal(t, 0, countType[Desertion](other))
}

func TestDealer_OtherCards_Desertions_PresentWhenConfigured(t *testing.T) {
	d := NewDealer(DeckConfig{Desertions: 2, ConstructionCards: 1})
	other := d.OtherCards(0)

	assert.Equal(t, 2, countType[Desertion](other))
}

// --- ConstructionCards minimum ---

func TestDealer_OtherCards_ConstructionCards_ZeroSkipsValueOne(t *testing.T) {
	// ConstructionCards=0 means no value-1 cards; only the fixed values 2-9 are added
	d := NewDealer(DeckConfig{ConstructionCards: 0})
	other := d.OtherCards(0)

	// 0 value-1 cards + 8 fixed cards for values 2-9 = 8 each
	// gold also gets 3 extra copies at values 5, 7, 9 = 11 total
	assert.Equal(t, 11, countType[*gold](other), "gold")
	assert.Equal(t, 8, countType[*sword](other), "swords")
	assert.Equal(t, 8, countType[*arrow](other), "arrows")
	assert.Equal(t, 8, countType[*poison](other), "poisons")
}
