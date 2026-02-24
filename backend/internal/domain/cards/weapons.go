package cards

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// FieldChecker is the minimal field interface needed by weapons to validate
// whether the player's current field supports using them.
// board.Field satisfies this interface.
type FieldChecker interface {
	HasWarriorType(t types.WarriorType) bool
}

type Weapon interface {
	Card
	DamageAmount() int
	Type() types.WeaponType
	CanConstruct() bool
	MultiplierFactor(target Warrior) int
	// CanBeUsedWith returns true if the player's field supports this weapon.
	// Special weapons (Harpoon, BloodRain, etc.) always return true here since
	// they have their own game actions with dedicated validation.
	CanBeUsedWith(field FieldChecker) bool
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

func (s *sword) CanBeUsedWith(field FieldChecker) bool {
	return field.HasWarriorType(types.KnightWarriorType) ||
		field.HasWarriorType(types.DragonWarriorType) ||
		field.HasWarriorType(types.MercenaryWarriorType)
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

func (s *arrow) CanBeUsedWith(field FieldChecker) bool {
	return field.HasWarriorType(types.ArcherWarriorType) ||
		field.HasWarriorType(types.DragonWarriorType) ||
		field.HasWarriorType(types.MercenaryWarriorType)
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

func (s *poison) CanBeUsedWith(field FieldChecker) bool {
	return field.HasWarriorType(types.MageWarriorType) ||
		field.HasWarriorType(types.DragonWarriorType) ||
		field.HasWarriorType(types.MercenaryWarriorType)
}
