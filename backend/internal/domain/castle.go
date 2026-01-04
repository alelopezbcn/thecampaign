package domain

import "fmt"

type Castle struct {
	cards []Card
}

func NewCastle(card Card) (*Castle, error) {
	if card.Value != 1 {
		return nil, fmt.Errorf("invalid card for constructing the castle")
	}

	return &Castle{
		cards: []Card{},
	}, nil
}

func (c *Castle) Value() int {
	total := 0
	for _, card := range c.cards {
		total += card.Value
	}

	return total
}

func (c *Castle) AddResource(card Card) error {
	if !card.IsResource() {
		return fmt.Errorf("card is not resource")
	}

	c.cards = append(c.cards, card)
	return nil
}

func (c *Castle) String() string {
	return fmt.Sprintf("Castle: %v Gold coins", c.Value())
}
