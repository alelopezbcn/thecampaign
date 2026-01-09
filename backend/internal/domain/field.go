package domain

import (
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type field struct {
	cards             []ports.Card
	gameEndedObserver ports.FieldWithoutWarriorsObserver
}

func NewField(o ports.FieldWithoutWarriorsObserver) ports.Field {
	return &field{
		cards:             []ports.Card{},
		gameEndedObserver: o,
	}
}

func (h *field) AddCards(cards ...ports.Card) {
	h.cards = append(h.cards, cards...)
}

func (h *field) ShowCards() []ports.Card {
	return h.cards
}

func (h *field) GetCard(cardID string) (ports.Card, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *field) RemoveCard(card ports.Card) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}

	return false
}
