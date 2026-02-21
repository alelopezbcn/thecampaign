package board

import "github.com/alelopezbcn/thecampaign/internal/domain/ports"

type cemetery struct {
	corps []ports.Warrior
}

func newCemetery() *cemetery {
	return &cemetery{
		corps: []ports.Warrior{},
	}
}

func (c *cemetery) Count() int {
	return len(c.corps)
}

func (c *cemetery) AddCorp(w ports.Warrior) {
	c.corps = append(c.corps, w)
}

func (c *cemetery) GetLast() ports.Warrior {
	if c.Count() == 0 {
		return nil
	}

	return c.corps[len(c.corps)-1]
}

func (c *cemetery) Corps() []ports.Warrior {
	return c.corps
}
