package board

import "github.com/alelopezbcn/thecampaign/internal/domain/ports"

type discardPile struct {
	cards []ports.Card
}

func newDiscardPile() *discardPile {
	return &discardPile{
		cards: []ports.Card{},
	}
}

func (d *discardPile) Count() int {
	return len(d.cards)
}

func (d *discardPile) Empty() []ports.Card {
	c := d.cards
	d.cards = []ports.Card{}
	return c
}

func (d *discardPile) Discard(c ports.Card) {
	d.cards = append(d.cards, c)
}

func (d *discardPile) GetLast() ports.Card {
	if d.Count() == 0 {
		return nil
	}

	return d.cards[len(d.cards)-1]
}

func (d *discardPile) Cards() []ports.Card {
	return d.cards
}
