package cards

type weaponCardBase struct {
	damageAmount int
}

func (s *weaponCardBase) DamageAmount() int {
	return s.damageAmount
}
