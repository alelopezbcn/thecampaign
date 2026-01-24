package domain

import (
	"fmt"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

const MaxCastleResources = 25

type castle struct {
	id                       string
	isConstructed            bool
	initialCard              ports.Card
	resources                []ports.Resource
	castleCompletionObserver ports.CastleCompletionObserver
	player                   ports.Player
}

func newCastle(p ports.Player, o ports.CastleCompletionObserver) ports.Castle {
	return &castle{
		id:                       "castle_" + p.Name(),
		resources:                []ports.Resource{},
		player:                   p,
		castleCompletionObserver: o,
	}
}

func (c *castle) GetID() string {
	return c.id
}

func (c *castle) Construct(card ports.Card) error {
	if !c.isConstructed {
		switch valuableCard := card.(type) {
		case ports.Weapon:
			if valuableCard.DamageAmount() != 1 {
				return fmt.Errorf("invalid card for constructing the castle")
			}
		case ports.Resource:
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

func (c *castle) RemoveGold(position int) (ports.Resource, error) {
	if len(c.resources) == 0 {
		return nil, fmt.Errorf("no Resource cards to remove from castle")
	}

	if position < 1 || position > len(c.resources) {
		return nil, fmt.Errorf("invalid position %d for removing a Resource from castle", position)
	}

	// Create a copy of c.resources and shuffle it
	copied := make([]ports.Resource, len(c.resources))
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

func (c *castle) CanBeAttacked() bool {
	return c.IsConstructed() && c.ResourceCards() > 0
}

func (c *castle) String() string {
	return fmt.Sprintf("Castle: %v Gold coins (%d cards)",
		c.Value(), c.ResourceCards())
}

func (c *castle) addResource(card ports.Card) error {
	gold, ok := card.(ports.Resource)
	if !ok {
		return fmt.Errorf("cardBase is not gold")
	}

	c.resources = append(c.resources, gold)
	if c.Value() >= MaxCastleResources {
		c.castleCompletionObserver.OnCastleCompletion(c.player)
	}

	return nil
}
