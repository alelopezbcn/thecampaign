package gameevents

import "github.com/alelopezbcn/thecampaign/internal/domain/types"

type calmHandler struct{}

func (h *calmHandler) ExtraDrawCards() int                         { return 0 }
func (h *calmHandler) WeaponDamageModifier(_ types.WeaponType) int { return 0 }
func (h *calmHandler) ConstructionValueModifier() int              { return 0 }
func (h *calmHandler) TurnStartWarriorHPModifier() int             { return 0 }
func (h *calmHandler) Display() (string, string) {
	return "Calm", "No special effects this round"
}
