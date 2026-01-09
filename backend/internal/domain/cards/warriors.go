package cards

import (
	"errors"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type knightCard struct {
	warriorCardBase
}

func NewKnightCard(id string) ports.Warrior {
	return &knightCard{
		warriorCardBase: newWarriorCardBase(
			cardBase{
				id:   strings.ToUpper(id),
				name: "Knight",
			},
			attackableCardBase{
				health:     WarriorHealth,
				attackedBy: []ports.Weapon{},
			},
		),
	}
}
func (k *knightCard) Attack(t ports.Attackable, w ports.Weapon) error {
	if t == nil {
		return errors.New("target cannot be nil")
	}
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	_, ok := w.(*swordCard)
	if !ok {
		return errors.New("knight can only attack with sword")
	}

	multiplier := 1
	if _, ok = t.(*archerCard); ok {
		multiplier = 2
	}

	t.ReceiveDamage(w, multiplier)

	return nil
}

type archerCard struct {
	warriorCardBase
}

func NewArcherCard(id string) ports.Warrior {
	return &archerCard{
		warriorCardBase: newWarriorCardBase(
			cardBase{
				id:   strings.ToUpper(id),
				name: "Archer",
			},
			attackableCardBase{
				health:     WarriorHealth,
				attackedBy: []ports.Weapon{},
			},
		),
	}
}
func (a *archerCard) Attack(t ports.Attackable, w ports.Weapon) error {
	if t == nil {
		return errors.New("target cannot be nil")
	}
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	_, ok := w.(*arrowCard)
	if !ok {
		return errors.New("archer can only attack with arrow")
	}

	multiplier := 1
	if _, ok = t.(*mageCard); ok {
		multiplier = 2
	}

	t.ReceiveDamage(w, multiplier)

	return nil
}

type mageCard struct {
	warriorCardBase
}

func NewMageCard(id string) ports.Warrior {
	return &mageCard{
		warriorCardBase: newWarriorCardBase(
			cardBase{
				id:   strings.ToUpper(id),
				name: "Mage",
			},
			attackableCardBase{
				health:     WarriorHealth,
				attackedBy: []ports.Weapon{},
			},
		),
	}
}
func (m *mageCard) Attack(t ports.Attackable, w ports.Weapon) error {
	if t == nil {
		return errors.New("target cannot be nil")
	}
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	_, ok := w.(*poisonCard)
	if !ok {
		return errors.New("mage can only attack with poison")
	}

	multiplier := 1
	if _, ok = t.(*knightCard); ok {
		multiplier = 2
	}

	t.ReceiveDamage(w, multiplier)

	return nil
}

type dragonCard struct {
	warriorCardBase
}

func NewDragonCard(id string) ports.Warrior {
	return &dragonCard{
		warriorCardBase: newWarriorCardBase(
			cardBase{
				id:   strings.ToUpper(id),
				name: "Dragon",
			},
			attackableCardBase{
				health:     DragonHealth,
				attackedBy: []ports.Weapon{},
			},
		),
	}
}
func (d *dragonCard) Attack(t ports.Attackable, w ports.Weapon) error {
	if t == nil {
		return errors.New("target cannot be nil")
	}
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1

	switch w.(type) {
	case *swordCard:
		if _, ok := t.(*archerCard); ok {
			multiplier = 2
		}
	case *arrowCard:
		if _, ok := t.(*mageCard); ok {
			multiplier = 2
		}
	case *poisonCard:
		if _, ok := t.(*knightCard); ok {
			multiplier = 2
		}
	}

	t.ReceiveDamage(w, multiplier)

	return nil
}
