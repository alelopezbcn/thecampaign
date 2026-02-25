package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

func TestNewAmbush(t *testing.T) {
	a := NewAmbush("amb1")
	assert.Equal(t, "AMB1", a.GetID())
	assert.Equal(t, "Ambush", a.Name())
}

func TestAmbush_ImplementsInterface(t *testing.T) {
	a := NewAmbush("amb1")
	var _ Ambush = a
}

func TestAmbush_Effect_IsValid(t *testing.T) {
	a := NewAmbush("amb1")
	effect := a.Effect()
	valid := effect == types.AmbushEffectReflectDamage ||
		effect == types.AmbushEffectCancelAttack ||
		effect == types.AmbushEffectStealWeapon ||
		effect == types.AmbushEffectDrainLife ||
		effect == types.AmbushEffectInstantKill
	assert.True(t, valid, "effect %d is not a valid AmbushEffect", effect)
}

func TestRandomAmbushEffect_AllEffectsReachable(t *testing.T) {
	seen := map[types.AmbushEffect]bool{}
	for i := 0; i < 10000; i++ {
		seen[types.RandomAmbushEffect()] = true
	}
	assert.True(t, seen[types.AmbushEffectReflectDamage], "AmbushEffectReflectDamage never produced")
	assert.True(t, seen[types.AmbushEffectCancelAttack], "AmbushEffectCancelAttack never produced")
	assert.True(t, seen[types.AmbushEffectStealWeapon], "AmbushEffectStealWeapon never produced")
	assert.True(t, seen[types.AmbushEffectDrainLife], "AmbushEffectDrainLife never produced")
	assert.True(t, seen[types.AmbushEffectInstantKill], "AmbushEffectInstantKill never produced")
}
