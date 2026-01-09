package domain

import (
	"strings"
)

type Field interface {
	ShowCards() []Card
	GetCard(cardID string) (Card, bool)
	AddCards(cards ...Card)
	RemoveCard(card Card) bool
}

type field struct {
	cards             []Card
	gameEndedObserver FieldWithoutWarriorsObserver
}

func newField(o FieldWithoutWarriorsObserver) Field {
	return &field{
		cards:             []Card{},
		gameEndedObserver: o,
	}
}

func (h *field) AddCards(cards ...Card) {
	h.cards = append(h.cards, cards...)
}

func (h *field) ShowCards() []Card {
	return h.cards
}

func (h *field) GetCard(cardID string) (Card, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *field) RemoveCard(card Card) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}

	return false
}
