package gameevents

import "github.com/alelopezbcn/thecampaign/internal/domain/types"

const bloodlustHealAmount = 2

// bloodlustHandler restores HP to a warrior each time it kills an enemy.
type bloodlustHandler struct{}

func (h *bloodlustHandler) ExtraDrawCards() int                         { return 0 }
func (h *bloodlustHandler) WeaponDamageModifier(_ types.WeaponType) int { return 0 }
func (h *bloodlustHandler) ConstructionValueModifier() int              { return 0 }
func (h *bloodlustHandler) TurnStartWarriorHPModifier() int             { return 0 }
func (h *bloodlustHandler) OnKillHealAmount() int                       { return bloodlustHealAmount }
func (h *bloodlustHandler) Display() (string, string) {
	return "Bloodlust", "Warriors restore 2 HP each time they defeat an enemy"
}
