package cards

import (
	"fmt"

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
	gold, err := castle.RemoveGold(position)
	if err != nil {
		return nil, err
	}

	return gold, nil
}
func (c *catapult) String() string {
	return fmt.Sprintf("%s (%s)", c.name, c.id)
}
