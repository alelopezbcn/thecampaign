package domain

import (
	"fmt"
	"strings"
)

type Thief interface {
	Card
}

type thiefCard struct {
	cardBase
}

func newThiefCard(id string) Thief {
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
