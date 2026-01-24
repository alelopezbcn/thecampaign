package gamestatus

import (
	"fmt"
)

type card struct {
	CardID   string
	CardType cardType
	Value    int
}

func newCard(cardID string, cardType cardType, value int) card {
	return card{
		CardID:   cardID,
		CardType: cardType,
		Value:    value,
	}
}

func (c card) String() string {
	return fmt.Sprintf("%s - %s (%d)", c.CardID, c.CardType.String(), c.Value)
}
