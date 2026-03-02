package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
)

type Card struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	SubType string `json:"sub_type,omitempty"`
	Color   string `json:"color"`
	Value   int    `json:"value,omitempty"`
}

func newCard(id string, ct CardType, value int) Card {
	return Card{
		ID:      id,
		Type:    ct.Name,
		SubType: ct.SubName,
		Color:   ct.Color,
		Value:   value,
	}
}

// CardType reconstructs the CardType from a Card's flattened fields.
// Primarily used in tests and comparisons.
func (c Card) CardType() CardType {
	return CardType{Name: c.Type, SubName: c.SubType, Color: c.Color}
}

func fromDomainCards(dcs []cards.Card) []Card {
	result := []Card{}
	for _, v := range dcs {
		result = append(result, fromDomainCard(v))
	}
	return result
}

func fromDomainCard(dc cards.Card) Card {
	cardID := dc.GetID()
	var ct CardType
	var value int

	switch c := dc.(type) {
	case cards.Warrior:
		ct = warriorCardTypes[c.Type()]
		value = c.Health()

	case cards.Weapon:
		ct = weaponCardTypes[c.Type()]
		value = c.DamageAmount()

	case cards.Resource:
		ct = CardTypeResource
		value = c.Value()

	default:
		ct = zeroValueCardTypes[dc.Name()]
	}

	return newCard(cardID, ct, value)
}
