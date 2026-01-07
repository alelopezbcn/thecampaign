package domain

import "fmt"

const MaxCastleResources = 25

type Castle struct {
	isConstructed            bool
	initialCard              iCard
	resources                []*goldCard
	castleCompletionObserver CastleCompletionObserver
	player                   *Player
}

func newCastle(p *Player, o CastleCompletionObserver) *Castle {
	return &Castle{
		resources:                []*goldCard{},
		player:                   p,
		castleCompletionObserver: o,
	}
}

func (c *Castle) Construct(card iCard) error {
	if !c.isConstructed {
		switch valuableCard := card.(type) {
		case weapon, *goldCard:
			if valuableCard.GetValue() != 1 {
				return fmt.Errorf("invalid card for constructing the castle")
			}
		default:
			return fmt.Errorf("invalid card type for constructing the castle")
		}

		c.isConstructed = true
		c.initialCard = card

		return nil
	}

	if err := c.addResource(card); err != nil {
		return err
	}

	return nil
}

func (c *Castle) IsConstructed() bool {
	return c.isConstructed
}

func (c *Castle) Value() int {
	total := 0
	for _, card := range c.resources {
		total += card.GetValue()
	}

	return total
}

func (c *Castle) addResource(card iCard) error {
	gold, ok := card.(*goldCard)
	if !ok {
		return fmt.Errorf("card is not gold")
	}

	c.resources = append(c.resources, gold)
	if c.Value() >= MaxCastleResources {
		c.castleCompletionObserver.OnCastleCompletion(c.player)
	}

	return nil
}

func (c *Castle) String() string {
	return fmt.Sprintf("Castle: %v Gold coins", c.Value())
}
