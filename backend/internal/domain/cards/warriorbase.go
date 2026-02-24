package cards

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

const (
	warriorMaxHealth = 20
	dragonMaxHealth  = 20
)

type warriorBase struct {
	*cardBase
	*attackableBase
	protectedBy         SpecialPower
	warriorType         types.WarriorType
	WarriorDeadObserver WarriorDeadObserver
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
func (w *warriorBase) Type() types.WarriorType {
	return w.warriorType
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
	w.WarriorDeadObserver.OnWarriorDead(w)
}
