package domain

import (
	"fmt"
	"strings"
)

type Catapult interface {
	Card
	Attack(castle *Castle, position int) (Resource, error)
}

type catapultCard struct {
	cardBase
}

func newCatapultCard(id string) Catapult {
	return &catapultCard{
		cardBase: cardBase{
			id:   strings.ToUpper(id),
			name: "Catapult",
		},
	}
}
func (c *catapultCard) Attack(castle *Castle, position int) (Resource, error) {
	gold, err := castle.RemoveGold(position)
	if err != nil {
		return nil, err
	}

	return gold, nil
}
func (c *catapultCard) String() string {
	return fmt.Sprintf("%s (%s)", c.name, c.id)
}
