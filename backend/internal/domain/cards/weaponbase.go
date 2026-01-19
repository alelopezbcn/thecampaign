package cards

import "github.com/alelopezbcn/thecampaign/internal/domain/ports"

type weaponBase struct {
	damageAmount int
	weaponType   ports.WeaponType
}

func newWeaponBase(damageAmount int, weaponType ports.WeaponType) *weaponBase {
	return &weaponBase{
		damageAmount: damageAmount,
		weaponType:   weaponType,
	}
}

func (s *weaponBase) DamageAmount() int {
	return s.damageAmount
}

func (s *weaponBase) Type() ports.WeaponType {
	return s.weaponType
}
