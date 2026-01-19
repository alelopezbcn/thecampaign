package cards

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type warriorBase struct {
	*cardBase
	*attackableBase
	protectedBy         ports.SpecialPower
	warriorType         ports.WarriorType
	WarriorDeadObserver ports.WarriorDeadObserver
}

func newWarriorBase(cardBase *cardBase, attackableCardBase *attackableBase,
	warriorType ports.WarriorType) *warriorBase {
	return &warriorBase{
		cardBase:       cardBase,
		attackableBase: attackableCardBase,
		warriorType:    warriorType,
	}
}

func (w *warriorBase) ReceiveDamage(weaponCard ports.Weapon, multiplier int) (isDefeated bool) {
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
func (w *warriorBase) Attack(_ ports.Attackable, _ ports.Weapon) error {
	return errors.New("should be implemented by concrete warrior types")
}
func (w *warriorBase) Protect(powerCard ports.SpecialPower) error {
	if w.protectedBy != nil {
		return errors.New("warrior already protected")
	}
	w.protectedBy = powerCard

	return nil
}
func (w *warriorBase) IsProtected() (bool, ports.Card) {
	if w.protectedBy != nil {
		return true, w.protectedBy
	}
	return false, nil
}
func (w *warriorBase) Heal(sp ports.SpecialPower) {
	w.health = WarriorMaxHealth
	w.attackedBy = append(w.attackedBy, sp)
	for _, a := range w.attackedBy {
		a.GetCardToBeDiscardedObserver().OnCardToBeDiscarded(a)
	}
	w.attackedBy = []ports.Weapon{}
}
func (w *warriorBase) InstantKill(sp ports.SpecialPower) {
	if w.protectedBy != nil {
		w.protectedBy.Destroyed()
		return
	}
	w.health = 0
	w.attackedBy = append(w.attackedBy, sp)
	w.dead()
}
func (w *warriorBase) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", w.name, w.id))
	if w.health > 0 {
		sb.WriteString(fmt.Sprintf(" - Health: %d", w.health))
	}
	if w.attackedBy != nil && len(w.attackedBy) > 0 {
		for _, card := range w.attackedBy {
			sb.WriteString(fmt.Sprintf("\n     * %s", card.String()))
		}
	}
	return sb.String()
}
func (w *warriorBase) AddWarriorDeadObserver(o ports.WarriorDeadObserver) {
	w.WarriorDeadObserver = o
}
func (w *warriorBase) Type() ports.WarriorType {
	return w.warriorType
}
func (w *warriorBase) IsDamaged() bool {
	return w.health < WarriorMaxHealth
}
func (w *warriorBase) dead() {
	for _, a := range w.attackedBy {
		a.GetCardToBeDiscardedObserver().OnCardToBeDiscarded(a)
	}
	w.attackedBy = []ports.Weapon{}
	w.WarriorDeadObserver.OnWarriorDead(w)
}
