package cards

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

const SpecialPowerDamage = 10

type specialPower struct {
	*cardBase
	*attackableBase
	*weaponBase
}

func NewSpecialPower(id string) ports.SpecialPower {
	return &specialPower{
		cardBase:       newCardBase(id, "Special Power"),
		attackableBase: newAttackableBase(SpecialPowerMaxHealth),
		weaponBase:     newWeaponBase(SpecialPowerDamage, types.SpecialPowerWeaponType),
	}
}
func (s *specialPower) BeAttacked(w ports.Weapon) error {
	if w == nil {
		return errors.New("weapon cannot be nil")
	}

	multiplier := 1
	s.ReceiveDamage(w, multiplier)

	return nil
}
func (s *specialPower) Use(usedBy ports.Warrior, target ports.Warrior) error {
	if _, ok := usedBy.(*dragon); ok {
		return errors.New("special power action not allowed to be used by Dragon")
	}

	switch usedBy.(type) {
	case *knight:
		if _, ok := target.(*dragon); ok {
			return errors.New("dragon cannot be protected")
		}
		if err := target.Protect(s); err != nil {
			return err
		}
	case *archer:
		target.InstantKill(s)
	case *mage:
		if _, ok := target.(*dragon); ok {
			return errors.New("dragon cannot be healed")
		}
		target.Heal(s)
	default:
		return errors.New("special power action not allowed for this warrior type")
	}

	return nil
}
func (s *specialPower) Destroyed() {
	for _, a := range s.attackedBy {
		a.GetCardMovedToPileObserver().OnCardMovedToPile(a)
	}
	s.attackedBy = []ports.Weapon{}
	s.cardMovedToPileObserver.OnCardMovedToPile(s)
}
func (s *specialPower) ReceiveDamage(w ports.Weapon, _ int) (isDefeated bool) {
	s.health -= w.DamageAmount()
	s.attackedBy = append(s.attackedBy, w)

	if s.health <= 0 {
		s.Destroyed()
		return true
	}

	return false
}
func (s *specialPower) String() string {
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
