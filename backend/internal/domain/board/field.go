package board

import (
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// FieldReader — read-only field access
type FieldReader interface {
	Warriors() []cards.Warrior
	GetWarrior(cardID string) (cards.Warrior, bool)
	HasWarriorType(t types.WarriorType) bool
	SlotCards() []cards.Card
}

// FieldMutator — field mutation
type FieldMutator interface {
	AddWarriors(cards ...cards.Warrior)
	RemoveWarrior(card cards.Warrior) bool
	SetSlotCard(c cards.Card)
	RemoveSlotCard(c cards.Card)
}

// Field composes read and write access
type Field interface {
	FieldReader
	FieldMutator
}

type FieldWithoutWarriorsObserver interface {
	OnFieldWithoutWarriors(playerName string)
}

type field struct {
	playerName        string
	cards             []cards.Warrior
	slotCards         []cards.Card
	gameEndedObserver FieldWithoutWarriorsObserver
}

func NewField(playerName string, o FieldWithoutWarriorsObserver) *field {
	return &field{
		playerName:        playerName,
		cards:             []cards.Warrior{},
		gameEndedObserver: o,
	}
}

func (h *field) AddWarriors(cards ...cards.Warrior) {
	h.cards = append(h.cards, cards...)
}

func (h *field) Warriors() []cards.Warrior {
	result := make([]cards.Warrior, len(h.cards))
	copy(result, h.cards)
	return result
}

func (h *field) GetWarrior(cardID string) (cards.Warrior, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *field) RemoveWarrior(card cards.Warrior) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			if len(h.cards) == 0 {
				h.gameEndedObserver.OnFieldWithoutWarriors(h.playerName)
			}
			return true
		}
	}

	return false
}

func (h *field) HasWarriorType(t types.WarriorType) bool {
	for _, w := range h.cards {
		if w.Type() == t {
			return true
		}
	}
	return false
}

func (h *field) SlotCards() []cards.Card {
	result := make([]cards.Card, len(h.slotCards))
	copy(result, h.slotCards)
	return result
}

func (h *field) SetSlotCard(c cards.Card) {
	h.slotCards = append(h.slotCards, c)
}

func (h *field) RemoveSlotCard(c cards.Card) {
	for i, s := range h.slotCards {
		if s.GetID() == c.GetID() {
			h.slotCards = append(h.slotCards[:i], h.slotCards[i+1:]...)
			return
		}
	}
}

// GetFieldSlotCard returns the first slot card of type T, or the zero value and false.
func GetFieldSlotCard[T any](f FieldReader) (T, bool) {
	for _, c := range f.SlotCards() {
		if card, ok := c.(T); ok {
			return card, true
		}
	}
	var zero T
	return zero, false
}

// HasFieldSlotCard reports whether the field has a slot card of type T.
func HasFieldSlotCard[T any](f FieldReader) bool {
	_, ok := GetFieldSlotCard[T](f)
	return ok
}
