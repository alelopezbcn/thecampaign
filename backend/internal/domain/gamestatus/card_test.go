package gamestatus

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/stretchr/testify/assert"
)

func TestFromDomainCard_Warrior(t *testing.T) {
	tests := []struct {
		name     string
		card     cards.Card
		wantType CardType
	}{
		{"Knight", cards.NewKnight("W1"), CardTypeKnight},
		{"Archer", cards.NewArcher("W1"), CardTypeArcher},
		{"Mage", cards.NewMage("W1"), CardTypeMage},
		{"Dragon", cards.NewDragon("W1"), CardTypeDragon},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fromDomainCard(tt.card)

			assert.Equal(t, "W1", c.CardID)
			assert.Equal(t, tt.wantType, c.CardType)
			assert.Equal(t, 20, c.Value) // warriorMaxHealth / dragonMaxHealth = 20
		})
	}
}

func TestFromDomainCard_Weapon(t *testing.T) {
	tests := []struct {
		name      string
		card      cards.Card
		wantType  CardType
		wantValue int
	}{
		{"Sword", cards.NewSword("WP1", 7), CardTypeSword, 7},
		{"Arrow", cards.NewArrow("WP1", 7), CardTypeArrow, 7},
		{"Poison", cards.NewPoison("WP1", 7), CardTypePoison, 7},
		{"SpecialPower", cards.NewSpecialPower("WP1"), CardTypeSpecialPower, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fromDomainCard(tt.card)

			assert.Equal(t, "WP1", c.CardID)
			assert.Equal(t, tt.wantType, c.CardType)
			assert.Equal(t, tt.wantValue, c.Value)
		})
	}
}

func TestFromDomainCard_Resource(t *testing.T) {
	resource := cards.NewGold("G1", 5)

	c := fromDomainCard(resource)

	assert.Equal(t, "G1", c.CardID)
	assert.Equal(t, CardTypeResource, c.CardType)
	assert.Equal(t, 5, c.Value)
}

func TestFromDomainCard_Spy(t *testing.T) {
	spy := cards.NewSpy("SPY1")

	c := fromDomainCard(spy)

	assert.Equal(t, "SPY1", c.CardID)
	assert.Equal(t, CardTypeSpy, c.CardType)
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Thief(t *testing.T) {
	thief := cards.NewThief("THIEF1")

	c := fromDomainCard(thief)

	assert.Equal(t, "THIEF1", c.CardID)
	assert.Equal(t, CardTypeThief, c.CardType)
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Catapult(t *testing.T) {
	catapult := cards.NewCatapultCard("CAT1")

	c := fromDomainCard(catapult)

	assert.Equal(t, "CAT1", c.CardID)
	assert.Equal(t, CardTypeCatapult, c.CardType)
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCards(t *testing.T) {
	warrior := cards.NewKnight("K1")
	resource := cards.NewGold("G1", 3)

	result := fromDomainCards([]cards.Card{warrior, resource})

	assert.Len(t, result, 2)
	assert.Equal(t, "K1", result[0].CardID)
	assert.Equal(t, CardTypeKnight, result[0].CardType)
	assert.Equal(t, "G1", result[1].CardID)
	assert.Equal(t, CardTypeResource, result[1].CardType)
}

func TestFromDomainCards_Empty(t *testing.T) {
	result := fromDomainCards([]cards.Card{})

	assert.Empty(t, result)
}

