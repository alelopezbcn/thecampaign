package domain

import "fmt"

type Castle struct {
	cards []iCard
}

func NewCastle(card iCard) (*Castle, error) {
	switch c := card.(type) {
	case *arrowCard, *poisonCard, *swordCard:
		if c.GetValue() != 1 {
			return nil, fmt.Errorf("invalid card for constructing the castle")
		}
	default:
		return nil, fmt.Errorf("invalid card type for constructing the castle")
	}

	return &Castle{
		cards: []iCard{},
	}, nil
}

func (c *Castle) Value() int {
	total := 0
	for _, card := range c.cards {
		total += card.GetValue()
	}

	return total
}

func (c *Castle) AddResource(card iCard) error {
	if !card.IsResource() {
		return fmt.Errorf("card is not resource")
	}

	c.cards = append(c.cards, card)
	return nil
}

func (c *Castle) String() string {
	return fmt.Sprintf("Castle: %v Gold coins", c.Value())
}
