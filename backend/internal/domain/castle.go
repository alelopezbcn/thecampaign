package domain

import (
	"fmt"
	"math/rand"
)

const MaxCastleResources = 25

type Castle interface {
	Construct(card Card) error
	IsConstructed() bool
	Value() int
	ResourceCards() int
	RemoveGold(position int) (Resource, error)
	String() string
}

type castle struct {
	isConstructed            bool
	initialCard              Card
	resources                []Resource
	castleCompletionObserver CastleCompletionObserver
	player                   Player
}

func newCastle(p Player, o CastleCompletionObserver) Castle {
	return &castle{
		resources:                []Resource{},
		player:                   p,
		castleCompletionObserver: o,
	}
}

func (c *castle) Construct(card Card) error {
	if !c.isConstructed {
		switch valuableCard := card.(type) {
		case Weapon:
			if valuableCard.DamageAmount() != 1 {
				return fmt.Errorf("invalid card for constructing the castle")
			}
		case Resource:
			if valuableCard.Value() != 1 {
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

func (c *castle) IsConstructed() bool {
	return c.isConstructed
}

func (c *castle) Value() int {
	total := 0
	for _, card := range c.resources {
		total += card.Value()
	}

	return total
}

func (c *castle) ResourceCards() int {
	return len(c.resources)
}

func (c *castle) RemoveGold(position int) (Resource, error) {
	if len(c.resources) == 0 {
		return nil, fmt.Errorf("no Resource cards to remove from castle")
	}

	if position < 1 || position > len(c.resources) {
		return nil, fmt.Errorf("invalid position %d for removing a Resource from castle", position)
	}

	// Create a copy of c.resources and shuffle it
	copied := make([]Resource, len(c.resources))
	copy(copied, c.resources)
	// Shuffle copied slice
	for i := range copied {
		j := i + rand.Intn(len(copied)-i)
		copied[i], copied[j] = copied[j], copied[i]
	}

	removedCard := copied[position-1]
	for i, r := range c.resources {
		if r.GetID() == removedCard.GetID() {
			c.resources = append(c.resources[:i], c.resources[i+1:]...)
			return r, nil
		}
	}

	return nil, fmt.Errorf("failed to remove Resource cardBase from castle")
}

func (c *castle) String() string {
	return fmt.Sprintf("Castle: %v Gold coins (%d cards)",
		c.Value(), c.ResourceCards())
}

func (c *castle) addResource(card Card) error {
	gold, ok := card.(Resource)
	if !ok {
		return fmt.Errorf("cardBase is not gold")
	}

	c.resources = append(c.resources, gold)
	if c.Value() >= MaxCastleResources {
		c.castleCompletionObserver.OnCastleCompletion(c.player)
	}

	return nil
}
