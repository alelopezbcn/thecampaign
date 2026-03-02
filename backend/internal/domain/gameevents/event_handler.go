package gameevents

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// EventHandler defines the interface for applying global event effects to game actions.
// Each event type has its own implementation file.
type EventHandler interface {
	// ExtraDrawCards returns the number of additional cards to draw during the draw phase (0 if none).
	ExtraDrawCards() int
	// WeaponDamageModifier returns the flat damage modifier for the given weapon type (0 if unaffected).
	WeaponDamageModifier(weaponType types.WeaponType) int
	// ConstructionValueModifier returns the flat modifier applied to each resource card's value during construction (0 if none).
	ConstructionValueModifier() int
	// TurnStartWarriorHPModifier returns the HP modifier applied to the active player's warriors at turn start (0 if none).
	TurnStartWarriorHPModifier() int
	// Display returns the event's display name and a human-readable description of its effect.
	Display() (name, description string)
}

// NewHandler creates the EventHandler for the given active event.
func NewHandler(event types.ActiveEvent) EventHandler {
	switch event.Type {
	case types.EventTypeCurse:
		return &curseHandler{excludedWeapon: event.CurseExcludedWeapon, modifier: event.CurseModifier}
	case types.EventTypeHarvest:
		return &harvestHandler{modifier: event.HarvestModifier}
	case types.EventTypePlague:
		return &plagueHandler{modifier: event.PlagueModifier}
	case types.EventTypeAbundance:
		return &abundanceHandler{}
	default:
		return &calmHandler{}
	}
}
