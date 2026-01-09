package cards

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type catapultCard struct {
	cardBase
}

func NewCatapultCard(id string) ports.Catapult {
	return &catapultCard{
		cardBase: cardBase{
			id:   strings.ToUpper(id),
			name: "Catapult",
		},
	}
}
func (c *catapultCard) Attack(castle ports.Castle, position int) (ports.Resource, error) {
	gold, err := castle.RemoveGold(position)
	if err != nil {
		return nil, err
	}

	return gold, nil
}
func (c *catapultCard) String() string {
	return fmt.Sprintf("%s (%s)", c.name, c.id)
}
