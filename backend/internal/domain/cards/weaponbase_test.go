package cards

import "testing"

func TestNewWeaponBase(t *testing.T) {
	weapon := newWeaponBase(10)
	if weapon.damageAmount != 10 {
		t.Errorf("expected damage amount to be 10, got %d", weapon.damageAmount)
	}
}

func TestWeaponBase_DamageAmount(t *testing.T) {
	weapon := newWeaponBase(5)
	if weapon.DamageAmount() != 5 {
		t.Errorf("expected damage amount to be 5, got %d", weapon.DamageAmount())
	}
}
