package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
)

type Card struct {
	CardID   string   `json:"card_id"`
	CardType CardType `json:"card_type"`
	Value    int      `json:"value"`
}

func newCard(cardID string, cardType CardType, value int) Card {
	return Card{
		CardID:   cardID,
		CardType: cardType,
		Value:    value,
	}
}

func fromDomainCards(dcs []cards.Card) []Card {
	cards := []Card{}
	for _, v := range dcs {
		cards = append(cards, fromDomainCard(v))
	}

	return cards
}

func fromDomainCard(dc cards.Card) Card {
	cardID := dc.GetID()
	var aCardType CardType
	var value int

	switch c := dc.(type) {
	case cards.Warrior:
		aCardType = warriorCardTypes[c.Type()]
		value = c.Health()

	case cards.Weapon:
		aCardType = weaponCardTypes[c.Type()]
		value = c.DamageAmount()

	case cards.Resource:
		aCardType = CardTypeResource
		value = c.Value()

	default:
		aCardType = zeroValueCardTypes[dc.Name()]
	}

	return newCard(cardID, aCardType, value)
}
