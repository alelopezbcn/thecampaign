package domain

import (
	"fmt"
	"strings"
)

type Resource interface {
	Card
	Value() int
}

type resourceBase struct {
	value int
}

func (r *resourceBase) Value() int {
	return r.value
}

type goldCard struct {
	cardBase
	resourceBase
}

func newGoldCard(id string, value int) Resource {
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
