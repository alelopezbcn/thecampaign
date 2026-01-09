package cards

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type spyCard struct {
	cardBase
}

func NewSpyCard(id string) ports.Spy {
	return &spyCard{
		cardBase: cardBase{
			id:   strings.ToUpper(id),
			name: "Spy",
		},
	}
}
func (s *spyCard) String() string {
	return fmt.Sprintf("%s (%s)", s.name, s.id)
}
