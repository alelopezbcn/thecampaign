package gameevents_test

import (
	"strings"
	"testing"

	"github.com/alelopezbcn/thecampaign/internal/domain/gameevents"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
	"github.com/stretchr/testify/assert"
)

// ──────────────────────────────────────────────────────────────────────────────
// NewHandler factory
// ──────────────────────────────────────────────────────────────────────────────

func TestNewHandler_Calm(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeNone})
	name, _ := h.Display()
	assert.Equal(t, "Calm", name)
	assert.Equal(t, 0, h.ExtraDrawCards())
	assert.Equal(t, 0, h.WeaponDamageModifier(types.SwordWeaponType))
	assert.Equal(t, 0, h.ConstructionValueModifier())
	assert.Equal(t, 0, h.TurnStartWarriorHPModifier())
}

func TestNewHandler_Curse(t *testing.T) {
	event := types.ActiveEvent{
		Type:                types.EventTypeCurse,
		CurseExcludedWeapon: types.ArrowWeaponType,
		CurseModifier:       -2,
	}
	h := gameevents.NewHandler(event)
	name, _ := h.Display()
	assert.Equal(t, "Curse", name)
}

func TestNewHandler_Harvest(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypeHarvest, HarvestModifier: 3}
	h := gameevents.NewHandler(event)
	name, _ := h.Display()
	assert.Equal(t, "Bountiful Harvest", name)
	assert.Equal(t, 3, h.ConstructionValueModifier())
}

func TestNewHandler_Plague(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypePlague, PlagueModifier: -1}
	h := gameevents.NewHandler(event)
	name, _ := h.Display()
	assert.Equal(t, "Plague", name)
	assert.Equal(t, -1, h.TurnStartWarriorHPModifier())
}

func TestNewHandler_Abundance(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeAbundance})
	name, _ := h.Display()
	assert.Equal(t, "Abundance", name)
	assert.Equal(t, 1, h.ExtraDrawCards())
}

// ──────────────────────────────────────────────────────────────────────────────
// calmHandler
// ──────────────────────────────────────────────────────────────────────────────

func TestCalmHandler_AllMethodsReturnZero(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{})
	assert.Equal(t, 0, h.ExtraDrawCards())
	assert.Equal(t, 0, h.WeaponDamageModifier(types.SwordWeaponType))
	assert.Equal(t, 0, h.ConstructionValueModifier())
	assert.Equal(t, 0, h.TurnStartWarriorHPModifier())
	assert.Equal(t, 0, h.OnKillHealAmount())
}

func TestCalmHandler_Display(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{})
	name, desc := h.Display()
	assert.Equal(t, "Calm", name)
	assert.NotEmpty(t, desc)
}

// ──────────────────────────────────────────────────────────────────────────────
// curseHandler
// ──────────────────────────────────────────────────────────────────────────────

func TestCurseHandler_AffectedWeaponsReturnModifier(t *testing.T) {
	// Arrow is excluded; Sword and Poison are affected
	event := types.ActiveEvent{
		Type:                types.EventTypeCurse,
		CurseExcludedWeapon: types.ArrowWeaponType,
		CurseModifier:       -2,
	}
	h := gameevents.NewHandler(event)
	assert.Equal(t, -2, h.WeaponDamageModifier(types.SwordWeaponType))
	assert.Equal(t, -2, h.WeaponDamageModifier(types.PoisonWeaponType))
}

func TestCurseHandler_ExcludedWeaponReturnsZero(t *testing.T) {
	event := types.ActiveEvent{
		Type:                types.EventTypeCurse,
		CurseExcludedWeapon: types.ArrowWeaponType,
		CurseModifier:       -2,
	}
	h := gameevents.NewHandler(event)
	assert.Equal(t, 0, h.WeaponDamageModifier(types.ArrowWeaponType))
}

func TestCurseHandler_NonCurseWeaponsReturnZero(t *testing.T) {
	event := types.ActiveEvent{
		Type:                types.EventTypeCurse,
		CurseExcludedWeapon: types.SwordWeaponType,
		CurseModifier:       3,
	}
	h := gameevents.NewHandler(event)
	assert.Equal(t, 0, h.WeaponDamageModifier(types.HarpoonWeaponType))
	assert.Equal(t, 0, h.WeaponDamageModifier(types.BloodRainWeaponType))
	assert.Equal(t, 0, h.WeaponDamageModifier(types.SpecialPowerWeaponType))
}

func TestCurseHandler_OtherMethodsReturnZero(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypeCurse, CurseModifier: 1}
	h := gameevents.NewHandler(event)
	assert.Equal(t, 0, h.ExtraDrawCards())
	assert.Equal(t, 0, h.ConstructionValueModifier())
	assert.Equal(t, 0, h.TurnStartWarriorHPModifier())
	assert.Equal(t, 0, h.OnKillHealAmount())
}

func TestCurseHandler_Display_NegativeModifier(t *testing.T) {
	event := types.ActiveEvent{
		Type:                types.EventTypeCurse,
		CurseExcludedWeapon: types.ArrowWeaponType,
		CurseModifier:       -3,
	}
	h := gameevents.NewHandler(event)
	name, desc := h.Display()
	assert.Equal(t, "Curse", name)
	assert.Contains(t, desc, "-3")
	assert.Contains(t, desc, "Arrow")
}

func TestCurseHandler_Display_PositiveModifier(t *testing.T) {
	event := types.ActiveEvent{
		Type:                types.EventTypeCurse,
		CurseExcludedWeapon: types.SwordWeaponType,
		CurseModifier:       2,
	}
	h := gameevents.NewHandler(event)
	_, desc := h.Display()
	assert.True(t, strings.Contains(desc, "+2"), "positive modifier should have '+' prefix")
}

// ──────────────────────────────────────────────────────────────────────────────
// harvestHandler
// ──────────────────────────────────────────────────────────────────────────────

func TestHarvestHandler_ConstructionValueModifier(t *testing.T) {
	tests := []struct{ modifier int }{{3}, {-4}, {1}, {-1}}
	for _, tt := range tests {
		event := types.ActiveEvent{Type: types.EventTypeHarvest, HarvestModifier: tt.modifier}
		h := gameevents.NewHandler(event)
		assert.Equal(t, tt.modifier, h.ConstructionValueModifier())
	}
}

func TestHarvestHandler_OtherMethodsReturnZero(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypeHarvest, HarvestModifier: 2}
	h := gameevents.NewHandler(event)
	assert.Equal(t, 0, h.ExtraDrawCards())
	assert.Equal(t, 0, h.WeaponDamageModifier(types.SwordWeaponType))
	assert.Equal(t, 0, h.TurnStartWarriorHPModifier())
	assert.Equal(t, 0, h.OnKillHealAmount())
}

func TestHarvestHandler_Display_NegativeModifier(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypeHarvest, HarvestModifier: -3}
	h := gameevents.NewHandler(event)
	name, desc := h.Display()
	assert.Equal(t, "Poor Harvest", name)
	assert.Contains(t, desc, "-3")
}

func TestHarvestHandler_Display_PositiveModifier(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypeHarvest, HarvestModifier: 4}
	h := gameevents.NewHandler(event)
	_, desc := h.Display()
	assert.True(t, strings.Contains(desc, "+4"), "positive modifier should have '+' prefix")
}

// ──────────────────────────────────────────────────────────────────────────────
// plagueHandler
// ──────────────────────────────────────────────────────────────────────────────

func TestPlagueHandler_TurnStartWarriorHPModifier(t *testing.T) {
	tests := []struct{ modifier int }{{-3}, {-1}, {1}, {3}}
	for _, tt := range tests {
		event := types.ActiveEvent{Type: types.EventTypePlague, PlagueModifier: tt.modifier}
		h := gameevents.NewHandler(event)
		assert.Equal(t, tt.modifier, h.TurnStartWarriorHPModifier())
	}
}

func TestPlagueHandler_OtherMethodsReturnZero(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypePlague, PlagueModifier: -2}
	h := gameevents.NewHandler(event)
	assert.Equal(t, 0, h.ExtraDrawCards())
	assert.Equal(t, 0, h.WeaponDamageModifier(types.SwordWeaponType))
	assert.Equal(t, 0, h.ConstructionValueModifier())
	assert.Equal(t, 0, h.OnKillHealAmount())
}

func TestPlagueHandler_Display_NegativeModifier(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypePlague, PlagueModifier: -2}
	h := gameevents.NewHandler(event)
	name, desc := h.Display()
	assert.Equal(t, "Plague", name)
	assert.Contains(t, desc, "2")
	assert.Contains(t, desc, "lose")
}

func TestPlagueHandler_Display_PositiveModifier(t *testing.T) {
	event := types.ActiveEvent{Type: types.EventTypePlague, PlagueModifier: 1}
	h := gameevents.NewHandler(event)
	_, desc := h.Display()
	assert.Contains(t, desc, "gain")
	assert.Contains(t, desc, "1")
}

// ──────────────────────────────────────────────────────────────────────────────
// abundanceHandler
// ──────────────────────────────────────────────────────────────────────────────

func TestAbundanceHandler_ExtraDrawCards(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeAbundance})
	assert.Equal(t, 1, h.ExtraDrawCards())
}

func TestAbundanceHandler_OtherMethodsReturnZero(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeAbundance})
	assert.Equal(t, 0, h.WeaponDamageModifier(types.SwordWeaponType))
	assert.Equal(t, 0, h.ConstructionValueModifier())
	assert.Equal(t, 0, h.TurnStartWarriorHPModifier())
	assert.Equal(t, 0, h.OnKillHealAmount())
}

func TestAbundanceHandler_Display(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeAbundance})
	name, desc := h.Display()
	assert.Equal(t, "Abundance", name)
	assert.NotEmpty(t, desc)
}

// ──────────────────────────────────────────────────────────────────────────────
// bloodlustHandler
// ──────────────────────────────────────────────────────────────────────────────

func TestNewHandler_Bloodlust(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeBloodlust})
	name, _ := h.Display()
	assert.Equal(t, "Bloodlust", name)
	assert.Equal(t, 2, h.OnKillHealAmount())
}

func TestBloodlustHandler_OnKillHealAmount(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeBloodlust})
	assert.Equal(t, 2, h.OnKillHealAmount())
}

func TestBloodlustHandler_OtherMethodsReturnZero(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeBloodlust})
	assert.Equal(t, 0, h.ExtraDrawCards())
	assert.Equal(t, 0, h.WeaponDamageModifier(types.SwordWeaponType))
	assert.Equal(t, 0, h.ConstructionValueModifier())
	assert.Equal(t, 0, h.TurnStartWarriorHPModifier())
}

func TestBloodlustHandler_Display(t *testing.T) {
	h := gameevents.NewHandler(types.ActiveEvent{Type: types.EventTypeBloodlust})
	name, desc := h.Display()
	assert.Equal(t, "Bloodlust", name)
	assert.NotEmpty(t, desc)
}
