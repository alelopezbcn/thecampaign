package gameevents

import "github.com/alelopezbcn/thecampaign/internal/domain/types"

const championsBountyCards = 1

// championsBountyHandler grants the attacker a card when killing a warrior belonging to
// the enemy with the highest total field HP. Only active in FFA3/FFA5 game modes.
type championsBountyHandler struct{}

func (h *championsBountyHandler) ExtraDrawCards() int                         { return 0 }
func (h *championsBountyHandler) WeaponDamageModifier(_ types.WeaponType) int { return 0 }
func (h *championsBountyHandler) ConstructionValueModifier() int              { return 0 }
func (h *championsBountyHandler) TurnStartWarriorHPModifier() int             { return 0 }
func (h *championsBountyHandler) OnKillHealAmount() int                       { return 0 }
func (h *championsBountyHandler) OnKillBountyCards() int                      { return championsBountyCards }
func (h *championsBountyHandler) Display() (string, string) {
	return "Champion's Bounty", "Draw 1 card when a weapon attack kills a warrior belonging to the enemy with the highest total field HP (weapon attacks only — special powers do not trigger this)"
}
