package cards

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type goldCard struct {
	cardBase
	resourceBase
}

func NewGoldCard(id string, value int) ports.Resource {
	return &goldCard{
		cardBase: cardBase{
			id:   strings.ToUpper(id),
			name: "Gold Coin",
		},
		resourceBase: resourceBase{
			value: value,
		},
	}
}
func (g *goldCard) String() string {
	return fmt.Sprintf("%d %s (%s)", g.Value(), g.name, g.id)
}
