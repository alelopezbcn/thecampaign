package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

func TestNewWeaponBase(t *testing.T) {
	weapon := newWeaponBase(10, ports.ArrowType)
	if weapon.damageAmount != 10 {
		t.Errorf("expected damage amount to be 10, got %d", weapon.damageAmount)
	}
}

func TestWeaponBase_DamageAmount(t *testing.T) {
	weapon := newWeaponBase(5, ports.ArrowType)
	if weapon.DamageAmount() != 5 {
		t.Errorf("expected damage amount to be 5, got %d", weapon.DamageAmount())
	}
}
