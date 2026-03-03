// Package cards contains the implementation of the card types in the game, including Ambush cards.
package cards

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// Ambush is a trap card placed face-down in the player's field.
// When an enemy attacks any warrior in that field, the ambush triggers,
// applies its pre-determined effect, and discards itself.
type Ambush interface {
	Card
	Effect() types.AmbushEffect
}

type ambushCard struct {
	*cardBase
	effect types.AmbushEffect
}

func NewAmbush(id string) *ambushCard {
	return &ambushCard{
		cardBase: newCardBase(id, "Ambush"),
		effect:   types.RandomAmbushEffect(),
	}
}

func (a *ambushCard) Effect() types.AmbushEffect {
	return a.effect
}
