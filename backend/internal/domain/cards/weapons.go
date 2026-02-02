package cards

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type sword struct {
	*cardBase
	*weaponBase
}

func NewSword(id string, damageAmount int) ports.Sword {
	return &sword{
		cardBase:   newCardBase(id, "Sword"),
		weaponBase: newWeaponBase(damageAmount, types.SwordWeaponType),
	}
}
func (s *sword) MultiplierFactor(target ports.Warrior) int {
	if target.Type() == types.ArcherWarriorType {
		return 2
	}

	return 1
}

type arrow struct {
	*cardBase
	*weaponBase
}

func NewArrow(id string, damageAmount int) ports.Arrow {
	return &arrow{
		cardBase:   newCardBase(id, "Arrow"),
		weaponBase: newWeaponBase(damageAmount, types.ArrowWeaponType),
	}
}
func (s *arrow) MultiplierFactor(target ports.Warrior) int {
	if target.Type() == types.MageWarriorType {
		return 2
	}

	return 1
}

type poison struct {
	*cardBase
	*weaponBase
}

func NewPoison(id string, damageAmount int) ports.Poison {
	{
		return &poison{
			cardBase:   newCardBase(id, "Poison"),
			weaponBase: newWeaponBase(damageAmount, types.PoisonWeaponType),
		}
	}
}
func (s *poison) MultiplierFactor(target ports.Warrior) int {
	if target.Type() == types.KnightWarriorType {
		return 2
	}

	return 1
}
