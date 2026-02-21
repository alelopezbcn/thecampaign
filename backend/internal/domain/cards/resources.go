package cards

import (
	"fmt"
)

type Resource interface {
	Card
	Value() int
	CanConstruct() bool
}

type gold struct {
	*cardBase
	value int
}

func NewGold(id string, value int) *gold {
	return &gold{
		cardBase: newCardBase(id, "Gold Coin"),
		value:    value,
	}
}

func (g *gold) Value() int {
	return g.value
}

func (g *gold) CanConstruct() bool {
	return g.value == 1
}

func (g *gold) String() string {
	return fmt.Sprintf("%d %s (%s)", g.Value(), g.name, g.id)
}
