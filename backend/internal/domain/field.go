package domain

import (
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type field struct {
	cards             []ports.Warrior
	gameEndedObserver ports.FieldWithoutWarriorsObserver
}

// HasArcher implements ports.Field.
func (h *field) HasArcher() bool {
	for _, warriorInField := range h.cards {
		switch warriorInField.(type) {
		case ports.Archer:
			return true
		}
	}

	return false
}

// HasDragon implements ports.Field.
func (h *field) HasDragon() bool {
	for _, warriorInField := range h.cards {
		switch warriorInField.(type) {
		case ports.Dragon:
			return true
		}
	}

	return false
}

// HasKnight implements ports.Field.
func (h *field) HasKnight() bool {
	for _, warriorInField := range h.cards {
		switch warriorInField.(type) {
		case ports.Knight:
			return true
		}
	}

	return false
}

// HasMage implements ports.Field.
func (h *field) HasMage() bool {
	for _, warriorInField := range h.cards {
		switch warriorInField.(type) {
		case ports.Mage:
			return true
		}
	}

	return false
}

func NewField(o ports.FieldWithoutWarriorsObserver) ports.Field {
	return &field{
		cards:             []ports.Warrior{},
		gameEndedObserver: o,
	}
}

func (h *field) AddWarriors(cards ...ports.Warrior) {
	h.cards = append(h.cards, cards...)
}

func (h *field) Warriors() []ports.Warrior {
	return h.cards
}

func (h *field) GetWarrior(cardID string) (ports.Warrior, bool) {
	for _, c := range h.cards {
		if strings.ToLower(c.GetID()) == strings.TrimSpace(strings.ToLower(cardID)) {
			return c, true
		}
	}

	return nil, false
}

func (h *field) RemoveWarrior(card ports.Warrior) bool {
	for i, c := range h.cards {
		if c.GetID() == card.GetID() {
			h.cards = append(h.cards[:i], h.cards[i+1:]...)
			return true
		}
	}

	return false
}
