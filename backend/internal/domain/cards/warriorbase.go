package cards

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

const (
	warriorMaxHealth   = 20
	dragonMaxHealth    = 20
	mercenaryMaxHealth = 15
)

type warriorBase struct {
	*cardBase
	*attackableBase
	protectedBy         SpecialPower
	warriorType         types.WarriorType
	WarriorDeadObserver WarriorDeadObserver
	self                Warrior // concrete outer type, preserved through cemetery/resurrection
}

func newWarriorBase(cardBase *cardBase, attackableCardBase *attackableBase,
	warriorType types.WarriorType) *warriorBase {
	return &warriorBase{
		cardBase:       cardBase,
		attackableBase: attackableCardBase,
		warriorType:    warriorType,
	}
}

func (w *warriorBase) ReceiveDamage(weaponCard Weapon, multiplier int) (isDefeated bool) {
	if w.protectedBy != nil {
		if w.protectedBy.ReceiveDamage(weaponCard, 1) {
			w.protectedBy = nil
		}
		return false
	}

	amount := weaponCard.DamageAmount() * multiplier
	w.health -= amount
	w.attackedBy = append(w.attackedBy, weaponCard)

	if w.health <= 0 {
		w.dead()
		return true
	}

	return false
}
func (w *warriorBase) BeAttacked(_ Weapon) error {
	return errors.New("should be implemented by concrete warrior types")
}

func (w *warriorBase) Protect(powerCard SpecialPower) error {
	if w.protectedBy != nil {
		return errors.New("warrior already protected")
	}
	w.protectedBy = powerCard

	return nil
}
func (w *warriorBase) IsProtected() (bool, SpecialPower) {
	if w.protectedBy != nil {
		return true, w.protectedBy
	}
	return false, nil
}
func (w *warriorBase) Heal(sp SpecialPower) {
	w.health = warriorMaxHealth
	w.attackedBy = append(w.attackedBy, sp)
	for _, a := range w.attackedBy {
		a.GetCardMovedToPileObserver().OnCardMovedToPile(a)
	}
	w.attackedBy = []Weapon{}
}

// HealToMax restores the warrior to full health without requiring a SpecialPower card.
// Used by the Ambush Drain Life effect. Mercenary overrides this.
func (w *warriorBase) HealToMax() {
	w.health = warriorMaxHealth
}

// HealBy increases the warrior's health by amount, with no cap.
// Used by the Ambush Drain Life effect so the heal mirrors the damage dealt.
func (w *warriorBase) HealBy(amount int) {
	w.health += amount
}

// KillByAmbush instantly kills the warrior, bypassing Dragon immunity.
// Used by the Ambush Instant Kill effect.
func (w *warriorBase) KillByAmbush() {
	if w.protectedBy != nil {
		w.protectedBy.Destroyed()
		w.protectedBy = nil
		return
	}
	w.health = 0
	w.dead()
}

func (w *warriorBase) InstantKill(sp SpecialPower) {
	if w.protectedBy != nil {
		w.protectedBy.Destroyed()
		w.protectedBy = nil
		return
	}
	w.health = 0
	w.attackedBy = append(w.attackedBy, sp)
	w.dead()
}
func (w *warriorBase) String() string {
	return strings.TrimSpace(fmt.Sprintf("%s (%d)", w.warriorType, w.Health()))
}
func (w *warriorBase) AddWarriorDeadObserver(o WarriorDeadObserver) {
	w.WarriorDeadObserver = o
}

// setSelf stores a reference to the concrete warrior type that embeds this base.
// Must be called in each concrete constructor so that the cemetery/resurrection
// cycle preserves the concrete type when OnWarriorDead is called.
func (w *warriorBase) setSelf(self Warrior) {
	w.self = self
}
func (w *warriorBase) Type() types.WarriorType {
	return w.warriorType
}

func (w *warriorBase) CanUseWeapon(wep types.WeaponType) bool {
	switch wep {
	case types.SwordWeaponType:
		return w.warriorType == types.KnightWarriorType || w.warriorType == types.DragonWarriorType || w.warriorType == types.MercenaryWarriorType
	case types.ArrowWeaponType:
		return w.warriorType == types.ArcherWarriorType || w.warriorType == types.DragonWarriorType || w.warriorType == types.MercenaryWarriorType
	case types.PoisonWeaponType:
		return w.warriorType == types.MageWarriorType || w.warriorType == types.DragonWarriorType || w.warriorType == types.MercenaryWarriorType
	default:
		return true
	}
}
func (w *warriorBase) IsDamaged() bool {
	return w.health < warriorMaxHealth
}

func (w *warriorBase) Resurrect() {
	w.health = warriorMaxHealth
	w.attackedBy = []Weapon{}
	w.protectedBy = nil
}

func (w *warriorBase) dead() {
	for _, a := range w.attackedBy {
		a.GetCardMovedToPileObserver().OnCardMovedToPile(a)
	}
	w.attackedBy = []Weapon{}
	// Use w.self so the cemetery stores the concrete type (*knight, *archer, etc.)
	// rather than *warriorBase, preserving BeAttacked dispatch after resurrection.
	w.WarriorDeadObserver.OnWarriorDead(w.self)
}
