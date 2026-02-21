package board

import (
	"errors"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

const MaxCardsInHand = 7

var ErrHandLimitExceeded = errors.New("hand limit exceeded")

type hand struct {
	cards []ports.Card
}

func NewHand() *hand {
	return &hand{
		cards: []ports.Card{},
	}
}

func (h *hand) AddCards(cards ...ports.Card) error {
	if len(h.cards)+len(cards) > MaxCardsInHand {
		return ErrHandLimitExceeded
	}

	h.cards = append(h.cards, cards...)

	return nil
}

func (h *hand) ShowCards() []ports.Card {
	return h.cards
}

func (h *hand) GetCard(cardID string) (ports.Card, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *hand) RemoveCard(card ports.Card) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}

	return false
}

func (h *hand) CanAddCards(count int) bool {
	return len(h.cards)+count <= MaxCardsInHand
}

func (h *hand) Count() int {
	return len(h.cards)
}
