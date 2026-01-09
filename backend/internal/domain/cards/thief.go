package cards

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type thiefCard struct {
	cardBase
}

func NewThiefCard(id string) ports.Thief {
	return &thiefCard{
		cardBase: cardBase{
			id:   strings.ToUpper(id),
			name: "Thief",
		},
	}
}
func (t *thiefCard) String() string {
	return fmt.Sprintf("%s (%s)", t.name, t.id)
}
