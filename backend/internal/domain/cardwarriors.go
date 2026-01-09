package domain

import (
	"errors"
	"strings"
)

type knightCard struct {
	warriorCardBase
}

func newKnightCard(id string) Warrior {
	return &knightCard{
		warriorCardBase: warriorCardBase{
			cardBase: cardBase{
				id:   strings.ToUpper(id),
				name: "Knight",
			},
			attackableCardBase: attackableCardBase{
				health:     WarriorHealth,
				attackedBy: []Weapon{},
			},
		},
	}
}
func (k *knightCard) Attack(t Attackable, w Weapon) error {
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

func newArcherCard(id string) Warrior {
	return &archerCard{
		warriorCardBase: warriorCardBase{
			cardBase: cardBase{
				id:   strings.ToUpper(id),
				name: "Archer",
			},
			attackableCardBase: attackableCardBase{
				health:     WarriorHealth,
				attackedBy: []Weapon{},
			},
		},
	}
}
func (a *archerCard) Attack(t Attackable, w Weapon) error {
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

func newMageCard(id string) Warrior {
	return &mageCard{
		warriorCardBase: warriorCardBase{
			cardBase: cardBase{
				id:   strings.ToUpper(id),
				name: "Mage",
			},
			attackableCardBase: attackableCardBase{
				health:     WarriorHealth,
				attackedBy: []Weapon{},
			},
		},
	}
}
func (m *mageCard) Attack(t Attackable, w Weapon) error {
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

func newDragonCard(id string) Warrior {
	return &dragonCard{
		warriorCardBase: warriorCardBase{
			cardBase: cardBase{
				id:   strings.ToUpper(id),
				name: "Dragon",
			},
			attackableCardBase: attackableCardBase{
				health:     DragonHealth,
				attackedBy: []Weapon{},
			},
		},
	}
}
func (d *dragonCard) Attack(t Attackable, w Weapon) error {
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
