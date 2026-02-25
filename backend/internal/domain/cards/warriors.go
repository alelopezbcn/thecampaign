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
	HealToMax()
	HealBy(amount int)
	InstantKill(sp SpecialPower)
	KillByAmbush()
	AddWarriorDeadObserver(o WarriorDeadObserver)
	Type() types.WarriorType
	IsDamaged() bool
	Resurrect()
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

type Dragon Warrior

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

type Mercenary interface {
	Warrior
	IsMercenaryCard() bool
}

type mercenary struct {
	*warriorBase
}

func NewMercenary(id string) *mercenary {
	return &mercenary{
		warriorBase: newWarriorBase(
			newCardBase(id, "Mercenary"),
			newAttackableBase(mercenaryMaxHealth),
			types.MercenaryWarriorType,
		),
	}
}

func (m *mercenary) IsMercenaryCard() bool { return true }

func (m *mercenary) BeAttacked(w Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1
	m.ReceiveDamage(w, multiplier)

	return nil
}

func (m *mercenary) IsDamaged() bool {
	return m.health < mercenaryMaxHealth
}

func (m *mercenary) Heal(sp SpecialPower) {
	m.health = mercenaryMaxHealth
	m.attackedBy = append(m.attackedBy, sp)
	for _, a := range m.attackedBy {
		a.GetCardMovedToPileObserver().OnCardMovedToPile(a)
	}
	m.attackedBy = []Weapon{}
}

func (m *mercenary) HealToMax() {
	m.health = mercenaryMaxHealth
}

func (m *mercenary) HealBy(amount int) {
	m.health += amount
}

func (m *mercenary) Resurrect() {
	m.health = mercenaryMaxHealth
	m.attackedBy = []Weapon{}
	m.protectedBy = nil
}
