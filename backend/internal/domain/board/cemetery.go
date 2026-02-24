package board

import (
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
)

type Cemetery interface {
	Count() int
	AddCorp(cards.Warrior)
	GetLast() cards.Warrior
	Corps() []cards.Warrior
	RemoveRandom() cards.Warrior
}

type WarriorMovedToCemeteryObserver interface {
	OnWarriorMovedToCemetery(card cards.Warrior)
}

type cemetery struct {
	corps []cards.Warrior
}

func NewCemetery() *cemetery {
	return &cemetery{
		corps: []cards.Warrior{},
	}
}

func (c *cemetery) Count() int {
	return len(c.corps)
}

func (c *cemetery) AddCorp(w cards.Warrior) {
	c.corps = append(c.corps, w)
}

func (c *cemetery) GetLast() cards.Warrior {
	if c.Count() == 0 {
		return nil
	}

	return c.corps[len(c.corps)-1]
}

func (c *cemetery) Corps() []cards.Warrior {
	return c.corps
}

func (c *cemetery) RemoveRandom() cards.Warrior {
	if len(c.corps) == 0 {
		return nil
	}
	idx := rand.Intn(len(c.corps))
	w := c.corps[idx]
	c.corps = append(c.corps[:idx], c.corps[idx+1:]...)
	return w
}
