package cards

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type spy struct {
	*cardBase
}

func NewSpy(id string) ports.Spy {
	return &spy{
		cardBase: newCardBase(id, "Spy"),
	}
}

func (s *spy) CanSpy() {}

func (s *spy) String() string {
	return fmt.Sprintf("%s (%s)", s.name, s.id)
}
