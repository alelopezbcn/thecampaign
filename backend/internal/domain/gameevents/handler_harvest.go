package gameevents

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// harvestHandler modifies the effective value of resource cards during construction.
type harvestHandler struct {
	modifier int // [-4,+4] excl. 0
}

func (h *harvestHandler) ExtraDrawCards() int                         { return 0 }
func (h *harvestHandler) WeaponDamageModifier(_ types.WeaponType) int { return 0 }
func (h *harvestHandler) TurnStartWarriorHPModifier() int             { return 0 }
func (h *harvestHandler) OnKillHealAmount() int                       { return 0 }
func (h *harvestHandler) OnKillBountyCards() int                      { return 0 }
func (h *harvestHandler) ConstructionValueModifier() int              { return h.modifier }

func (h *harvestHandler) Display() (string, string) {
	if h.modifier > 0 {
		return "Bountiful Harvest", fmt.Sprintf(
			"Resources contribute +%d value to castle construction this round", h.modifier,
		)
	}
	return "Poor Harvest", fmt.Sprintf(
		"Resources contribute %d value to castle construction this round", h.modifier,
	)
}
