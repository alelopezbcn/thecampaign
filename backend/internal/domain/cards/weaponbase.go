package cards

import (
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

func (s *weaponBase) CanConstruct() bool {
	return s.DamageAmount() == 1
}
