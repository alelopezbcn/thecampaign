package domain

import (
	"strings"
)

type field struct {
	cards             []Card
	gameEndedObserver CastleCompletionObserver
}

func newField(o CastleCompletionObserver) *field {
	return &field{
		cards:             []Card{},
		gameEndedObserver: o,
	}
}

func (h *field) addCards(cards ...Card) {
	h.cards = append(h.cards, cards...)
}

func (h *field) showCards() []Card {
	return h.cards
}

func (h *field) getCard(cardID string) (Card, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *field) removeCard(card Card) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}

	return false
}
