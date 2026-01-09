package cards

type weaponBase struct {
	damageAmount int
}

func newWeaponBase(damageAmount int) *weaponBase {
	return &weaponBase{
		damageAmount: damageAmount,
	}
}

func (s *weaponBase) DamageAmount() int {
	return s.damageAmount
}
