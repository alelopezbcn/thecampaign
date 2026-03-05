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
	CanUseWeapon(wep types.WeaponType) bool
	Kills() int
	AddKill()
}

type knight struct {
	*warriorBase
}

func NewKnight(id string) *knight {
	k := &knight{
		warriorBase: newWarriorBase(
			newCardBase(id, "Knight"),
			newAttackableBase(warriorMaxHealth),
			types.KnightWarriorType,
		),
	}
	k.setSelf(k)
	return k
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
	a := &archer{
		warriorBase: newWarriorBase(
			newCardBase(id, "Archer"),
			newAttackableBase(warriorMaxHealth),
			types.ArcherWarriorType,
		),
	}
	a.setSelf(a)
	return a
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
	m := &mage{
		warriorBase: newWarriorBase(
			newCardBase(id, "Mage"),
			newAttackableBase(warriorMaxHealth),
			types.MageWarriorType,
		),
	}
	m.setSelf(m)
	return m
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
	d := &dragon{
		warriorBase: newWarriorBase(
			newCardBase(id, "Dragon"),
			newAttackableBase(dragonMaxHealth),
			types.DragonWarriorType,
		),
	}
	d.setSelf(d)
	return d
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
	m := &mercenary{
		warriorBase: newWarriorBase(
			newCardBase(id, "Mercenary"),
			newAttackableBase(mercenaryMaxHealth),
			types.MercenaryWarriorType,
		),
	}
	m.setSelf(m)
	return m
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
