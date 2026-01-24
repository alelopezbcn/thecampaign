package gamestatus

import (
	"fmt"
)

type Card struct {
	CardID   string
	CardType CardType
	Value    int
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
