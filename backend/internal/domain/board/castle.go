// Package board contains the implementation of the game board, including the Castle, Field, Deck, and DiscardPile.
package board

import (
	"fmt"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
)

// CastleReader — read-only castle access
type CastleReader interface {
	IsConstructed() bool
	IsProtected() bool
	Value() int
	ResourceCardsCount() int
	ResourceCards() []cards.Resource
	CanBeAttacked() bool
	ResourcesToWin() int
}

// CastleMutator — castle mutation
type CastleMutator interface {
	Construct(card cards.Card) error
	RemoveGold(position int) (cards.Resource, error)
	SetProtection(f cards.Fortress)
	ConsumeProtection() cards.Card
	// Reset clears all castle state and returns every card that was in it
	// (initial construction card, resources, protection) for the caller to discard.
	Reset() []cards.Card
}

// Castle composes read and write access
type Castle interface {
	CastleReader
	CastleMutator
}

type CastleCompletionObserver interface {
	OnCastleCompletion(playerName string)
}

type castle struct {
	id                       string
	isConstructed            bool
	initialCard              cards.Card
	resources                []cards.Resource
	protection               cards.Fortress
	castleCompletionObserver CastleCompletionObserver
	player                   Player
	resourcesToWin           int
}

func NewCastle(resourcesToWin int, p Player, o CastleCompletionObserver) *castle {
	return &castle{
		id:                       "castle_" + p.Name(),
		resources:                []cards.Resource{},
		player:                   p,
		castleCompletionObserver: o,
		resourcesToWin:           resourcesToWin,
	}
}

func (c *castle) GetID() string {
	return c.id
}

func (c *castle) Construct(card cards.Card) error {
	if !c.isConstructed {
		switch valuableCard := card.(type) {
		case cards.Weapon:
			if valuableCard.DamageAmount() != 1 {
				return fmt.Errorf("damage != 1: invalid card for constructing the castle")
			}
		case cards.Resource:
			if valuableCard.Value() != 1 {
				return fmt.Errorf("gold != 1: invalid card for constructing the castle")
			}
		default:
			return fmt.Errorf("invalid card type for constructing the castle: %T", card)
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

func (c *castle) IsProtected() bool {
	return c.protection != nil
}

func (c *castle) SetProtection(f cards.Fortress) {
	c.protection = f
}

func (c *castle) ConsumeProtection() cards.Card {
	f := c.protection
	c.protection = nil
	return f
}

func (c *castle) Value() int {
	total := 0
	for _, card := range c.resources {
		total += card.Value()
	}

	return total
}

func (c *castle) ResourceCardsCount() int {
	return len(c.resources)
}

func (c *castle) ResourceCards() []cards.Resource {
	return c.resources
}

func (c *castle) RemoveGold(position int) (cards.Resource, error) {
	if len(c.resources) == 0 {
		return nil, fmt.Errorf("no Resource cards to remove from castle")
	}

	if position < 1 || position > len(c.resources) {
		return nil, fmt.Errorf("invalid position %d for removing a Resource from castle", position)
	}

	// Create a copy of c.resources and shuffle it
	copied := make([]cards.Resource, len(c.resources))
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

	return nil, fmt.Errorf("failed to remove Resource card from castle")
}

func (c *castle) CanBeAttacked() bool {
	return c.IsConstructed() && (c.ResourceCardsCount() > 0 || c.IsProtected())
}

func (c *castle) ResourcesToWin() int {
	return c.resourcesToWin
}

func (c *castle) String() string {
	return fmt.Sprintf("Castle: %v Gold coins (%d cards)",
		c.Value(), c.ResourceCardsCount())
}

func (c *castle) Reset() []cards.Card {
	var discarded []cards.Card
	if c.initialCard != nil {
		discarded = append(discarded, c.initialCard)
	}
	for _, r := range c.resources {
		discarded = append(discarded, r)
	}
	if c.protection != nil {
		discarded = append(discarded, c.protection)
	}
	c.isConstructed = false
	c.initialCard = nil
	c.resources = []cards.Resource{}
	c.protection = nil
	return discarded
}

func (c *castle) addResource(card cards.Card) error {
	gold, ok := card.(cards.Resource)
	if !ok {
		return fmt.Errorf("card is not gold resource: %T", card)
	}

	c.resources = append(c.resources, gold)
	if c.Value() >= c.resourcesToWin {
		c.castleCompletionObserver.OnCastleCompletion(c.player.Name())
	}

	return nil
}
