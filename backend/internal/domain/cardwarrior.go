package domain

import (
	"errors"
	"fmt"
	"strings"
)

type Warrior interface {
	Card
	Attackable
	Attack(target Attackable, weapon Weapon) error
	ProtectedBy(powerCard SpecialPower)
	Heal()
	InstantKill()
	AddWarriorDeadObserver(o WarriorDeadObserver)
}

type warriorCardBase struct {
	cardBase
	attackableCardBase
	protectedBy         SpecialPower
	WarriorDeadObserver WarriorDeadObserver
}

func (w *warriorCardBase) AddWarriorDeadObserver(o WarriorDeadObserver) {
	w.WarriorDeadObserver = o
}
func (w *warriorCardBase) ReceiveDamage(weaponCard Weapon, multiplier int) (isDefeated bool) {
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
func (w *warriorCardBase) String() string {
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
func (w *warriorCardBase) Attack(_ Attackable, _ Weapon) error {
	return errors.New("should be implemented by concrete Warrior types")
}
func (w *warriorCardBase) ProtectedBy(powerCard SpecialPower) {
	w.protectedBy = powerCard
}
func (w *warriorCardBase) Heal() {
	w.health = WarriorHealth
	for _, a := range w.attackedBy {
		a.GetCardToBeDiscardedObserver().OnCardToBeDiscarded(a)
	}
	w.attackedBy = []Weapon{}
}
func (w *warriorCardBase) InstantKill() {
	w.dead()
}
func (w *warriorCardBase) dead() {
	for _, a := range w.attackedBy {
		a.GetCardToBeDiscardedObserver().OnCardToBeDiscarded(a)
	}

	w.attackedBy = []Weapon{}
	w.WarriorDeadObserver.OnWarriorDead(w)
}
