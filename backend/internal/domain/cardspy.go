package domain

import (
	"fmt"
	"strings"
)

type Spy interface {
	Card
}

type spyCard struct {
	cardBase
}

func newSpyCard(id string) Spy {
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
