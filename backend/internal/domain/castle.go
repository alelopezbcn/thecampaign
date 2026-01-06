package domain

import "fmt"

type Castle struct {
	cards []resource
}

func NewCastle(card iCard) (*Castle, error) {
	switch c := card.(type) {
	case weapon:
		if c.GetValue() != 1 {
			return nil, fmt.Errorf("invalid card for constructing the castle")
		}
	default:
		return nil, fmt.Errorf("invalid card type for constructing the castle")
	}

	return &Castle{
		cards: []resource{},
	}, nil
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
	return nil
}

func (c *Castle) String() string {
	return fmt.Sprintf("Castle: %v Gold coins", c.Value())
}
