package cards

import (
	"errors"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type knight struct {
	*warriorBase
}

func NewKnight(id string) ports.Warrior {
	return &knight{
		warriorBase: newWarriorBase(
			newCardBase(id, "Knight"),
			newAttackableBase(WarriorHealth),
		),
	}
}
func (k *knight) Attack(t ports.Attackable, w ports.Weapon) error {
	if t == nil {
		return errors.New("target cannot be nil")
	}
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	_, ok := w.(*sword)
	if !ok {
		return errors.New("knight can only attack with sword")
	}

	multiplier := 1
	if _, ok = t.(*archer); ok {
		multiplier = 2
	}

	t.ReceiveDamage(w, multiplier)

	return nil
}

type archer struct {
	*warriorBase
}

func NewArcher(id string) ports.Warrior {
	return &archer{
		warriorBase: newWarriorBase(
			newCardBase(id, "Archer"),
			newAttackableBase(WarriorHealth),
		),
	}
}
func (a *archer) Attack(t ports.Attackable, w ports.Weapon) error {
	if t == nil {
		return errors.New("target cannot be nil")
	}
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	_, ok := w.(*arrow)
	if !ok {
		return errors.New("archer can only attack with arrow")
	}

	multiplier := 1
	if _, ok = t.(*mage); ok {
		multiplier = 2
	}

	t.ReceiveDamage(w, multiplier)

	return nil
}

type mage struct {
	*warriorBase
}

func NewMage(id string) ports.Warrior {
	return &mage{
		warriorBase: newWarriorBase(
			newCardBase(id, "Mage"),
			newAttackableBase(WarriorHealth),
		),
	}
}
func (m *mage) Attack(t ports.Attackable, w ports.Weapon) error {
	if t == nil {
		return errors.New("target cannot be nil")
	}
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	_, ok := w.(*poison)
	if !ok {
		return errors.New("mage can only attack with poison")
	}

	multiplier := 1
	if _, ok = t.(*knight); ok {
		multiplier = 2
	}

	t.ReceiveDamage(w, multiplier)

	return nil
}

type dragon struct {
	*warriorBase
}

func NewDragon(id string) ports.Warrior {
	return &dragon{
		warriorBase: newWarriorBase(
			newCardBase(id, "Dragon"),
			newAttackableBase(DragonHealth),
		),
	}
}
func (d *dragon) Attack(t ports.Attackable, w ports.Weapon) error {
	if t == nil {
		return errors.New("target cannot be nil")
	}
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1

	switch w.(type) {
	case *sword:
		if _, ok := t.(*archer); ok {
			multiplier = 2
		}
	case *arrow:
		if _, ok := t.(*mage); ok {
			multiplier = 2
		}
	case *poison:
		if _, ok := t.(*knight); ok {
			multiplier = 2
		}
	}

	t.ReceiveDamage(w, multiplier)

	return nil
}
