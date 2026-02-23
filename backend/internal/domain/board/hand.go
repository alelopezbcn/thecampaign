package board

import (
	"errors"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
)

const MaxCardsInHand = 7

var ErrHandLimitExceeded = errors.New("hand limit exceeded")

// HandReader — read-only hand access
type HandReader interface {
	ShowCards() []cards.Card
	GetCard(cardID string) (cards.Card, bool)
	CanAddCards(count int) bool
	Count() int
}

// HandMutator — hand mutation
type HandMutator interface {
	AddCards(cards ...cards.Card) error
	RemoveCard(card cards.Card) bool
}

// Hand composes read and write access
type Hand interface {
	HandReader
	HandMutator
}

type hand struct {
	cards []cards.Card
}

func NewHand() *hand {
	return &hand{
		cards: []cards.Card{},
	}
}

func (h *hand) AddCards(cards ...cards.Card) error {
	if len(h.cards)+len(cards) > MaxCardsInHand {
		return ErrHandLimitExceeded
	}

	h.cards = append(h.cards, cards...)

	return nil
}

func (h *hand) ShowCards() []cards.Card {
	return h.cards
}

func (h *hand) GetCard(cardID string) (cards.Card, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *hand) RemoveCard(card cards.Card) bool {
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
