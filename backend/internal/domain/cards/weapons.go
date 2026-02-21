package cards

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type Weapon interface {
	Card
	DamageAmount() int
	Type() types.WeaponType
	CanConstruct() bool
	MultiplierFactor(target Warrior) int
	String() string
}

type sword struct {
	*cardBase
	*weaponBase
}

func NewSword(id string, damageAmount int) *sword {
	return &sword{
		cardBase:   newCardBase(id, "Sword"),
		weaponBase: newWeaponBase(damageAmount, types.SwordWeaponType),
	}
}
func (s *sword) MultiplierFactor(target Warrior) int {
	if target.Type() == types.ArcherWarriorType {
		return 2
	}

	return 1
}

type arrow struct {
	*cardBase
	*weaponBase
}

func NewArrow(id string, damageAmount int) *arrow {
	return &arrow{
		cardBase:   newCardBase(id, "Arrow"),
		weaponBase: newWeaponBase(damageAmount, types.ArrowWeaponType),
	}
}
func (s *arrow) MultiplierFactor(target Warrior) int {
	if target.Type() == types.MageWarriorType {
		return 2
	}

	return 1
}

type poison struct {
	*cardBase
	*weaponBase
}

func NewPoison(id string, damageAmount int) *poison {
	{
		return &poison{
			cardBase:   newCardBase(id, "Poison"),
			weaponBase: newWeaponBase(damageAmount, types.PoisonWeaponType),
		}
	}
}
func (s *poison) MultiplierFactor(target Warrior) int {
	if target.Type() == types.KnightWarriorType {
		return 2
	}

	return 1
}
