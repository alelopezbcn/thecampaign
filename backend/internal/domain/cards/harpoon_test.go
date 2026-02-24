package cards

import (
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

func TestNewHarpoon(t *testing.T) {
	h := NewHarpoon("h1")
	assert.Equal(t, "H1", h.GetID())
	assert.Equal(t, harpoonDamage, h.DamageAmount())
	assert.Equal(t, types.HarpoonWeaponType, h.Type())
	assert.Contains(t, h.String(), "Harpoon")
}

func TestHarpoon_CanBeUsedWith(t *testing.T) {
	h := NewHarpoon("h1")
	// Harpoon always returns true regardless of field composition
	assert.True(t, h.CanBeUsedWith(nil))
}

func TestHarpoon_MultiplierFactor(t *testing.T) {
	h := NewHarpoon("h1")
	assert.Equal(t, 1, h.MultiplierFactor(NewKnight("k1")))
	assert.Equal(t, 1, h.MultiplierFactor(NewDragon("d1")))
}

func TestHarpoon_CanConstruct(t *testing.T) {
	h := NewHarpoon("h1")
	assert.False(t, h.CanConstruct())
}

func TestHarpoon_Attack_NilTarget(t *testing.T) {
	h := NewHarpoon("h1")
	err := h.Attack(nil)
	assert.ErrorContains(t, err, "target cannot be nil")
}

func TestHarpoon_Attack_KillsDragon(t *testing.T) {
	h := NewHarpoon("h1")
	// harpoonDamage == dragonMaxHealth, so the dragon dies on hit.
	// dead() notifies observers, so both must be wired up.
	cardObs := &fakeCardObs{}
	h.AddCardMovedToPileObserver(cardObs)

	target := NewDragon("d1")
	deadObs := &fakeWarriorDeadObs{}
	target.AddWarriorDeadObserver(deadObs)

	err := h.Attack(target)

	assert.NoError(t, err)
	assert.LessOrEqual(t, target.Health(), 0)
	assert.Len(t, deadObs.called, 1)
}
