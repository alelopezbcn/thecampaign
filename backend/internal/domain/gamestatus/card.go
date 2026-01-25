package gamestatus

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
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

func (c Card) String() string {
	return fmt.Sprintf("%s - %s (%d)", c.CardID, c.CardType.String(), c.Value)
}

func FromWarriorCard(c ports.Warrior) Card {
	var aCardType CardType
	switch c.Type() {
	case types.ArcherWarriorType:
		aCardType = CardTypeArcher
	case types.MageWarriorType:
		aCardType = CardTypeMage
	case types.KnightWarriorType:
		aCardType = CardTypeKnight
	}

	return newCard(c.GetID(), aCardType, c.Health())
}
