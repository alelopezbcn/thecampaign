package cards

import (
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

func (t *thief) CanSteal() {}
