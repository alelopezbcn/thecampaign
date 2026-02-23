package cards

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type weaponBase struct {
	damageAmount int
	weaponType   types.WeaponType
}

func newWeaponBase(damageAmount int, weaponType types.WeaponType) *weaponBase {
	return &weaponBase{
		damageAmount: damageAmount,
		weaponType:   weaponType,
	}
}

func (s *weaponBase) DamageAmount() int {
	return s.damageAmount
}

func (s *weaponBase) Type() types.WeaponType {
	return s.weaponType
}

func (s *weaponBase) MultiplierFactor(_ Warrior) int {
	return 1
}

func (s *weaponBase) CanConstruct() bool {
	return s.DamageAmount() == 1
}

func (s *weaponBase) CanBeUsedWith(_ FieldChecker) bool {
	return true
}

func (s *weaponBase) String() string {
	return fmt.Sprintf("%s (%d)", s.weaponType, s.damageAmount)
}
