package gamestatus

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/stretchr/testify/assert"
)

func TestFromDomainCard_Warrior(t *testing.T) {
	tests := []struct {
		name      string
		card      cards.Card
		wantType  CardType
		wantValue int
	}{
		{"Knight", cards.NewKnight("W1"), CardTypeKnight, 20},
		{"Archer", cards.NewArcher("W1"), CardTypeArcher, 20},
		{"Mage", cards.NewMage("W1"), CardTypeMage, 20},
		{"Dragon", cards.NewDragon("W1"), CardTypeDragon, 20},
		{"Mercenary", cards.NewMercenary("W1"), CardTypeMercenary, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fromDomainCard(tt.card)

			assert.Equal(t, "W1", c.ID)
			assert.Equal(t, tt.wantType, c.CardType())
			assert.Equal(t, tt.wantValue, c.Value)
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

			assert.Equal(t, "WP1", c.ID)
			assert.Equal(t, tt.wantType, c.CardType())
			assert.Equal(t, tt.wantValue, c.Value)
		})
	}
}

func TestFromDomainCard_Resource(t *testing.T) {
	resource := cards.NewGold("G1", 5)

	c := fromDomainCard(resource)

	assert.Equal(t, "G1", c.ID)
	assert.Equal(t, CardTypeResource, c.CardType())
	assert.Equal(t, 5, c.Value)
}

func TestFromDomainCard_Spy(t *testing.T) {
	spy := cards.NewSpy("SPY1")

	c := fromDomainCard(spy)

	assert.Equal(t, "SPY1", c.ID)
	assert.Equal(t, CardTypeSpy, c.CardType())
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Thief(t *testing.T) {
	thief := cards.NewThief("THIEF1")

	c := fromDomainCard(thief)

	assert.Equal(t, "THIEF1", c.ID)
	assert.Equal(t, CardTypeThief, c.CardType())
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Catapult(t *testing.T) {
	catapult := cards.NewCatapultCard("CAT1")

	c := fromDomainCard(catapult)

	assert.Equal(t, "CAT1", c.ID)
	assert.Equal(t, CardTypeCatapult, c.CardType())
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Fortress(t *testing.T) {
	fortress := cards.NewFortress("FW1")

	c := fromDomainCard(fortress)

	assert.Equal(t, "FW1", c.ID)
	assert.Equal(t, CardTypeFortress, c.CardType())
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Harpoon(t *testing.T) {
	harpoon := cards.NewHarpoon("H1")

	c := fromDomainCard(harpoon)

	assert.Equal(t, "H1", c.ID)
	assert.Equal(t, CardTypeHarpoon, c.CardType())
	assert.Equal(t, 20, c.Value) // harpoonDamage = 20; Harpoon is a Weapon, not default
}

func TestFromDomainCard_BloodRain(t *testing.T) {
	bloodRain := cards.NewBloodRain("BR1")

	c := fromDomainCard(bloodRain)

	assert.Equal(t, "BR1", c.ID)
	assert.Equal(t, CardTypeBloodRain, c.CardType())
	assert.Equal(t, 4, c.Value) // bloodRainDamage = 4; BloodRain is a Weapon, not default
}

func TestFromDomainCard_Resurrection(t *testing.T) {
	resurrection := cards.NewResurrection("RES1")

	c := fromDomainCard(resurrection)

	assert.Equal(t, "RES1", c.ID)
	assert.Equal(t, CardTypeResurrection, c.CardType())
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Sabotage(t *testing.T) {
	sabotage := cards.NewSabotage("SAB1")

	c := fromDomainCard(sabotage)

	assert.Equal(t, "SAB1", c.ID)
	assert.Equal(t, CardTypeSabotage, c.CardType())
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCards(t *testing.T) {
	warrior := cards.NewKnight("K1")
	resource := cards.NewGold("G1", 3)

	result := fromDomainCards([]cards.Card{warrior, resource})

	assert.Len(t, result, 2)
	assert.Equal(t, "K1", result[0].ID)
	assert.Equal(t, CardTypeKnight, result[0].CardType())
	assert.Equal(t, "G1", result[1].ID)
	assert.Equal(t, CardTypeResource, result[1].CardType())
}

func TestFromDomainCards_Empty(t *testing.T) {
	result := fromDomainCards([]cards.Card{})

	assert.Empty(t, result)
}

func TestFromDomainCard_Treason(t *testing.T) {
	treason := cards.NewTreason("DES1")

	c := fromDomainCard(treason)

	assert.Equal(t, "DES1", c.ID)
	assert.Equal(t, CardTypeTreason, c.CardType())
	assert.Equal(t, 0, c.Value)
}

func TestFromDomainCard_Ambush(t *testing.T) {
	ambush := cards.NewAmbush("AMB1")

	c := fromDomainCard(ambush)

	assert.Equal(t, "AMB1", c.ID)
	assert.Equal(t, CardTypeAmbush, c.CardType())
	assert.Equal(t, 0, c.Value)
}
