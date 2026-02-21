package cards

import (
	"errors"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type Warrior interface {
	Card
	Attackable
	Protect(powerCard SpecialPower) error
	IsProtected() (bool, SpecialPower)
	Heal(powerCard SpecialPower)
	InstantKill(sp SpecialPower)
	AddWarriorDeadObserver(o WarriorDeadObserver)
	Type() types.WarriorType
	IsDamaged() bool
}

type knight struct {
	*warriorBase
}

func NewKnight(id string) *knight {
	return &knight{
		warriorBase: newWarriorBase(
			newCardBase(id, "Knight"),
			newAttackableBase(warriorMaxHealth),
			types.KnightWarriorType,
		),
	}
}
func (k *knight) BeAttacked(w Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	k.ReceiveDamage(w, w.MultiplierFactor(k))

	return nil
}

type archer struct {
	*warriorBase
}

func NewArcher(id string) *archer {
	return &archer{
		warriorBase: newWarriorBase(
			newCardBase(id, "Archer"),
			newAttackableBase(warriorMaxHealth),
			types.ArcherWarriorType,
		),
	}
}
func (a *archer) BeAttacked(w Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	a.ReceiveDamage(w, w.MultiplierFactor(a))

	return nil
}

type mage struct {
	*warriorBase
}

func NewMage(id string) *mage {
	return &mage{
		warriorBase: newWarriorBase(
			newCardBase(id, "Mage"),
			newAttackableBase(warriorMaxHealth),
			types.MageWarriorType,
		),
	}
}
func (m *mage) BeAttacked(w Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	m.ReceiveDamage(w, w.MultiplierFactor(m))

	return nil
}

type dragon struct {
	*warriorBase
}

func NewDragon(id string) *dragon {
	return &dragon{
		warriorBase: newWarriorBase(
			newCardBase(id, "Dragon"),
			newAttackableBase(dragonMaxHealth),
			types.DragonWarriorType,
		),
	}
}
func (d *dragon) BeAttacked(w Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1
	d.ReceiveDamage(w, multiplier)

	return nil
}
func (d *dragon) InstantKill(sp SpecialPower) {
	// Dragon cannot be instant killed
	d.health -= sp.DamageAmount()
	d.attackedBy = append(d.attackedBy, sp)

	if d.health <= 0 {
		d.dead()
	}
}
