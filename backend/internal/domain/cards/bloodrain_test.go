package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

func TestNewBloodRain(t *testing.T) {
	b := NewBloodRain("b1")
	assert.Equal(t, "B1", b.GetID())
	assert.Equal(t, bloodRainDamage, b.DamageAmount())
	assert.Equal(t, types.BloodRainWeaponType, b.Type())
	assert.Contains(t, b.String(), "Blood Rain")
}

func TestBloodRain_CanBeUsedWith(t *testing.T) {
	b := NewBloodRain("b1")
	// BloodRain always returns true regardless of field composition
	assert.True(t, b.CanBeUsedWith(nil))
}

func TestBloodRain_MultiplierFactor(t *testing.T) {
	b := NewBloodRain("b1")
	assert.Equal(t, 1, b.MultiplierFactor(NewKnight("k1")))
	assert.Equal(t, 1, b.MultiplierFactor(NewArcher("a1")))
}

func TestBloodRain_CanConstruct(t *testing.T) {
	b := NewBloodRain("b1")
	assert.False(t, b.CanConstruct())
}

func TestBloodRain_Attack_EmptyTargets(t *testing.T) {
	b := NewBloodRain("b1")
	err := b.Attack([]Warrior{})
	assert.ErrorContains(t, err, "targets cannot be empty")
}

func TestBloodRain_Attack_SingleTarget(t *testing.T) {
	b := NewBloodRain("b1")
	target := NewKnight("k1")

	err := b.Attack([]Warrior{target})

	assert.NoError(t, err)
	assert.Equal(t, warriorMaxHealth-bloodRainDamage, target.Health())
}

func TestBloodRain_Attack_MultipleTargets(t *testing.T) {
	b := NewBloodRain("b1")
	knight := NewKnight("k1")
	archer := NewArcher("a1")
	mage := NewMage("m1")

	err := b.Attack([]Warrior{knight, archer, mage})

	assert.NoError(t, err)
	assert.Equal(t, warriorMaxHealth-bloodRainDamage, knight.Health())
	assert.Equal(t, warriorMaxHealth-bloodRainDamage, archer.Health())
	assert.Equal(t, warriorMaxHealth-bloodRainDamage, mage.Health())
}
