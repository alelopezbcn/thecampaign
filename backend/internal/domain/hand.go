package domain

import (
	"errors"
	"strings"
)

const maxCardsInHand = 7

var ErrHandLimitExceeded = errors.New("hand limit exceeded")

type hand struct {
	cards []Card
}

func newHand() *hand {
	return &hand{
		cards: []Card{},
	}
}

func (h *hand) addCards(cards ...Card) error {
	if len(h.cards)+len(cards) > maxCardsInHand {
		return ErrHandLimitExceeded
	}

	h.cards = append(h.cards, cards...)

	return nil
}

func (h *hand) showCards() []Card {
	return h.cards
}

func (h *hand) getCard(cardID string) (Card, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *hand) removeCard(card Card) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}

	return false
}

func (h *hand) canAddCards(count int) bool {
	return len(h.cards)+count <= maxCardsInHand
}
