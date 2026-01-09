package domain

import (
	"errors"
	"fmt"
	"strings"
)

const SpecialPowerDamage = 10

type SpecialPower interface {
	Card
	Attackable
	Use(warriorCard Warrior, targetCard Attackable) error
}
type specialPowerCard struct {
	cardBase
	attackableCardBase
	weaponCardBase
}

func newSpecialPowerCard(id string) SpecialPower {
	return &specialPowerCard{
		cardBase: cardBase{
			id:   strings.ToUpper(id),
			name: "Special Power",
		},
		attackableCardBase: attackableCardBase{
			health:     SpecialPowerHealth,
			attackedBy: []Weapon{},
		},
		weaponCardBase: weaponCardBase{
			damageAmount: SpecialPowerDamage,
		},
	}
}
func (s *specialPowerCard) Use(w Warrior, t Attackable) error {
	if _, ok := w.(*dragonCard); ok {
		return errors.New("special power action not allowed to be used by Dragon")
	}

	if warriorTarget, ok := t.(Warrior); ok {
		switch warriorTarget.(type) {
		case *knightCard:
			warriorTarget.ProtectedBy(s)
		case *archerCard:
			warriorTarget.InstantKill()
		case *mageCard:
			warriorTarget.Heal()
		case *dragonCard:
			warriorTarget.ReceiveDamage(s, 1)
		default:
			return errors.New("special power action not allowed for this Warrior type")
		}
	}

	// TOOO: seguir por aca

	return nil
}
func (s *specialPowerCard) ReceiveDamage(w Weapon, _ int) (isDefeated bool) {
	s.health -= w.DamageAmount()
	s.attackedBy = append(s.attackedBy, w)

	if s.health <= 0 {
		for _, a := range s.attackedBy {
			a.GetCardToBeDiscardedObserver().OnCardToBeDiscarded(a)
		}
		s.cardToBeDiscardedObserver.OnCardToBeDiscarded(s)
		s.attackedBy = []Weapon{}

		return true
	}

	return false
}
func (s *specialPowerCard) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", s.name, s.id))
	if s.health > 0 {
		sb.WriteString(fmt.Sprintf(" - Health: %d", s.health))
	}
	if s.attackedBy != nil && len(s.attackedBy) > 0 {
		for _, card := range s.attackedBy {
			sb.WriteString(fmt.Sprintf("\n  * %s", card.String()))
		}
	}
	return sb.String()
}
