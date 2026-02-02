package cards

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type catapult struct {
	*cardBase
}

func NewCatapultCard(id string) ports.Catapult {
	return &catapult{
		cardBase: newCardBase(id, "Catapult"),
	}
}
func (c *catapult) Attack(castle ports.Castle, position int) (ports.Resource, error) {
	g, err := castle.RemoveGold(position)
	if err != nil {
		return nil, err
	}

	return g, nil
}
