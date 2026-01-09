package cards

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type gold struct {
	*cardBase
	*resourceBase
}

func NewGold(id string, value int) ports.Resource {
	return &gold{
		cardBase:     newCardBase(id, "Gold Coin"),
		resourceBase: newResourceBase(value),
	}
}
func (g *gold) String() string {
	return fmt.Sprintf("%d %s (%s)", g.Value(), g.name, g.id)
}
