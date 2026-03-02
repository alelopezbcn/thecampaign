package gameevents

import "github.com/alelopezbcn/thecampaign/internal/domain/types"

// abundanceHandler grants the active player one extra card during the draw phase.
type abundanceHandler struct{}

func (h *abundanceHandler) ExtraDrawCards() int                          { return 1 }
func (h *abundanceHandler) WeaponDamageModifier(_ types.WeaponType) int  { return 0 }
func (h *abundanceHandler) ConstructionValueModifier() int               { return 0 }
func (h *abundanceHandler) TurnStartWarriorHPModifier() int              { return 0 }
func (h *abundanceHandler) Display() (string, string) {
	return "Abundance", "Draw 1 extra card at the start of your turn"
}
