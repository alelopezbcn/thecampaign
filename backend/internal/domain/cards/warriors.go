package cards

import (
	"errors"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type knight struct {
	*warriorBase
}

func NewKnight(id string) ports.Knight {
	return &knight{
		warriorBase: newWarriorBase(
			newCardBase(id, "Knight"),
			newAttackableBase(WarriorMaxHealth),
			types.KnightWarriorType,
		),
	}
}
func (k *knight) BeAttacked(w ports.Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1
	if w.Type() == types.PoisonWeaponType {
		multiplier = 2
	}

	k.ReceiveDamage(w, multiplier)

	return nil
}

type archer struct {
	*warriorBase
}

func NewArcher(id string) ports.Archer {
	return &archer{
		warriorBase: newWarriorBase(
			newCardBase(id, "Archer"),
			newAttackableBase(WarriorMaxHealth),
			types.ArcherWarriorType,
		),
	}
}
func (a *archer) BeAttacked(w ports.Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1
	if w.Type() == types.SwordWeaponType {
		multiplier = 2
	}

	a.ReceiveDamage(w, multiplier)

	return nil
}

type mage struct {
	*warriorBase
}

func NewMage(id string) ports.Mage {
	return &mage{
		warriorBase: newWarriorBase(
			newCardBase(id, "Mage"),
			newAttackableBase(WarriorMaxHealth),
			types.MageWarriorType,
		),
	}
}
func (m *mage) BeAttacked(w ports.Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1
	if w.Type() == types.ArrowWeaponType {
		multiplier = 2
	}

	m.ReceiveDamage(w, multiplier)

	return nil
}

type dragon struct {
	*warriorBase
}

func NewDragon(id string) ports.Dragon {
	return &dragon{
		warriorBase: newWarriorBase(
			newCardBase(id, "Dragon"),
			newAttackableBase(DragonMaxHealth),
			types.DragonWarriorType,
		),
	}
}
func (d *dragon) BeAttacked(w ports.Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1
	d.ReceiveDamage(w, multiplier)

	return nil
}
func (d *dragon) InstantKill(sp ports.SpecialPower) {
	// Dragon cannot be instant killed
	d.health -= sp.DamageAmount()
	d.attackedBy = append(d.attackedBy, sp)

	if d.health <= 0 {
		d.dead()
	}
}
