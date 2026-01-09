package domain

import (
	"errors"
	"strings"
)

const maxCardsInHand = 7

var ErrHandLimitExceeded = errors.New("hand limit exceeded")

type Hand interface {
	ShowCards() []Card
	GetCard(cardID string) (Card, bool)
	AddCards(cards ...Card) error
	RemoveCard(card Card) bool
	CanAddCards(count int) bool
}

type hand struct {
	cards []Card
}

func newHand() Hand {
	return &hand{
		cards: []Card{},
	}
}

func (h *hand) AddCards(cards ...Card) error {
	if len(h.cards)+len(cards) > maxCardsInHand {
		return ErrHandLimitExceeded
	}

	h.cards = append(h.cards, cards...)

	return nil
}

func (h *hand) ShowCards() []Card {
	return h.cards
}

func (h *hand) GetCard(cardID string) (Card, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *hand) RemoveCard(card Card) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}

	return false
}

func (h *hand) CanAddCards(count int) bool {
	return len(h.cards)+count <= maxCardsInHand
}
