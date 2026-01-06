package domain

import (
	"errors"
	"strings"
)

const maxCardsInHand = 7

var ErrHandLimitExceeded = errors.New("hand limit exceeded")

type hand struct {
	cards []iCard
}

func newHand() *hand {
	return &hand{
		cards: []iCard{},
	}
}

func (h *hand) addCards(cards ...iCard) error {
	if len(h.cards)+len(cards) > maxCardsInHand {
		return ErrHandLimitExceeded
	}

	h.cards = append(h.cards, cards...)

	return nil
}

func (h *hand) showCards() []iCard {
	return h.cards
}

func (h *hand) getCard(cardID string) (iCard, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *hand) removeCard(card iCard) bool {
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
