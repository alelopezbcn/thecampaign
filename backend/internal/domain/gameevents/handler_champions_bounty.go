package gameevents

import "github.com/alelopezbcn/thecampaign/internal/domain/types"

const championsBountyCards = 2

// championsBountyHandler grants the attacker a card when killing a warrior belonging to
// the enemy with the highest total field HP. Only active in FFA3/FFA5 game modes.
type championsBountyHandler struct{}

func (h *championsBountyHandler) ExtraDrawCards() int                         { return 0 }
func (h *championsBountyHandler) WeaponDamageModifier(_ types.WeaponType) int { return 0 }
func (h *championsBountyHandler) ConstructionValueModifier() int              { return 0 }
func (h *championsBountyHandler) TurnStartWarriorHPModifier() int             { return 0 }
func (h *championsBountyHandler) OnKillHealAmount() int                       { return 0 }
func (h *championsBountyHandler) OnKillBountyCards() int                      { return championsBountyCards }
func (h *championsBountyHandler) OnHitBountyHeal() int                        { return 3 }
func (h *championsBountyHandler) Display() (string, string) {
	return "Champion's Bounty", "Draw 2 cards when killing a warrior from the enemy with the highest HP"
}
