package cards

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type thief struct {
	*cardBase
}

func NewThief(id string) ports.Thief {
	return &thief{
		cardBase: newCardBase(id, "Thief"),
	}
}
func (t *thief) String() string {
	return fmt.Sprintf("%s (%s)", t.name, t.id)
}
