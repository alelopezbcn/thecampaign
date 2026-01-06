package domain

import "fmt"

const MaxCastleResources = 25

type Castle struct {
	isConstructed     bool
	cards             []resource
	gameEndedObserver GameEndedObserver
}

func newCastle(o GameEndedObserver) *Castle {
	return &Castle{
		cards:             []resource{},
		gameEndedObserver: o,
	}
}

func (c *Castle) Construct(card iCard) error {
	switch c := card.(type) {
	case weapon:
		if c.GetValue() != 1 {
			return fmt.Errorf("invalid card for constructing the castle")
		}
	default:
		return fmt.Errorf("invalid card type for constructing the castle")
	}

	c.isConstructed = true
	return nil
}

func (c *Castle) IsConstructed() bool {
	return c.isConstructed
}

func (c *Castle) Value() int {
	total := 0
	for _, card := range c.cards {
		total += card.GetValue()
	}

	return total
}

func (c *Castle) AddResource(card resource) error {
	if card == nil {
		return fmt.Errorf("card is not resource")
	}

	c.cards = append(c.cards, card)
	if c.Value() >= MaxCastleResources {
		c.gameEndedObserver.OnGameEnded("Castle has reached maximum resources")
	}

	return nil
}

func (c *Castle) String() string {
	return fmt.Sprintf("Castle: %v Gold coins", c.Value())
}
