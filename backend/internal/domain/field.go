package domain

import (
	"strings"
)

type field struct {
	cards             []iCard
	gameEndedObserver CastleCompletionObserver
}

func newField(o CastleCompletionObserver) *field {
	return &field{
		cards:             []iCard{},
		gameEndedObserver: o,
	}
}

func (h *field) addCards(cards ...iCard) {
	h.cards = append(h.cards, cards...)
}

func (h *field) showCards() []iCard {
	return h.cards
}

func (h *field) getCard(cardID string) (iCard, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *field) removeCard(card iCard) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}

	return false
}
