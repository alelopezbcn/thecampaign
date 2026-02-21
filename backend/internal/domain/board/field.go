package board

import (
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type Field interface {
	Warriors() []cards.Warrior
	GetWarrior(cardID string) (cards.Warrior, bool)
	AddWarriors(cards ...cards.Warrior)
	RemoveWarrior(card cards.Warrior) bool
	HasArcher() bool
	HasMage() bool
	HasKnight() bool
	HasDragon() bool
}

type FieldWithoutWarriorsObserver interface {
	OnFieldWithoutWarriors(playerName string)
}

type field struct {
	playerName        string
	cards             []cards.Warrior
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
	return h.cards
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

func (h *field) HasArcher() bool {
	for _, warriorInField := range h.cards {
		switch warriorInField.Type() {
		case types.ArcherWarriorType:
			return true
		}
	}

	return false
}

func (h *field) HasDragon() bool {
	for _, warriorInField := range h.cards {
		switch warriorInField.Type() {
		case types.DragonWarriorType:
			return true
		}
	}

	return false
}

func (h *field) HasKnight() bool {
	for _, warriorInField := range h.cards {
		switch warriorInField.Type() {
		case types.KnightWarriorType:
			return true
		}
	}

	return false
}

func (h *field) HasMage() bool {
	for _, warriorInField := range h.cards {
		switch warriorInField.Type() {
		case types.MageWarriorType:
			return true
		}
	}

	return false
}
