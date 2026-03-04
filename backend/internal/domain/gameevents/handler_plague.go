package gameevents

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// plagueHandler applies an HP modifier to the active player's warriors at the start of each turn.
// Positive modifier heals; negative modifier damages (but never kills).
type plagueHandler struct {
	modifier int // [-3,+3] excl. 0
}

func (h *plagueHandler) ExtraDrawCards() int                         { return 0 }
func (h *plagueHandler) WeaponDamageModifier(_ types.WeaponType) int { return 0 }
func (h *plagueHandler) ConstructionValueModifier() int              { return 0 }
func (h *plagueHandler) TurnStartWarriorHPModifier() int             { return h.modifier }
func (h *plagueHandler) OnKillHealAmount() int                       { return 0 }

func (h *plagueHandler) Display() (string, string) {
	if h.modifier > 0 {
		return "Plague", fmt.Sprintf(
			"Your warriors gain %d HP at the start of your turn", h.modifier,
		)
	}
	return "Plague", fmt.Sprintf(
		"Your warriors lose %d HP at the start of your turn (cannot die from this)", -h.modifier,
	)
}
