package gameevents

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// curseHandler affects two of the three basic weapons (Sword, Arrow, Poison).
// The excluded weapon is unaffected; the other two get the damage modifier.
type curseHandler struct {
	excludedWeapon types.WeaponType
	modifier       int
}

func (h *curseHandler) ExtraDrawCards() int             { return 0 }
func (h *curseHandler) ConstructionValueModifier() int  { return 0 }
func (h *curseHandler) TurnStartWarriorHPModifier() int { return 0 }

func (h *curseHandler) WeaponDamageModifier(weaponType types.WeaponType) int {
	if weaponType == h.excludedWeapon {
		return 0
	}
	// Only the three basic weapons can be affected
	for _, w := range types.CurseWeapons {
		if weaponType == w {
			return h.modifier
		}
	}
	return 0
}

func (h *curseHandler) Display() (string, string) {
	sign := "+"
	if h.modifier < 0 {
		sign = ""
	}
	return "Curse", fmt.Sprintf(
		"All weapons except %s deal %s%d damage this round",
		h.excludedWeapon, sign, h.modifier,
	)
}
