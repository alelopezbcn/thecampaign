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
	// OnKillHealAmount returns the HP restored to the killing warrior when it defeats an enemy (0 if none).
	OnKillHealAmount() int
	// OnKillBountyCards returns the number of cards drawn when killing a warrior belonging to the
	// enemy with the highest total field HP (0 if none). Only active in FFA modes.
	OnKillBountyCards() int
	// OnHitBountyHeal returns the HP healed to a random field warrior when hitting (without killing)
	// a warrior belonging to the enemy with the highest total field HP (0 if none).
	// For blood rain: triggers whenever the target player is top enemy, regardless of kills.
	OnHitBountyHeal() int
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
	case types.EventTypeBloodlust:
		return &bloodlustHandler{}
	case types.EventTypeChampionsBounty:
		return &championsBountyHandler{}
	default:
		return &calmHandler{}
	}
}
