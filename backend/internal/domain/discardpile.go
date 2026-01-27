package domain

import "github.com/alelopezbcn/thecampaign/internal/domain/ports"

type discardPile struct {
	cards []ports.Card
}

func newDiscardPile() ports.DiscardPile {
	return &discardPile{
		cards: []ports.Card{},
	}
}

func (d *discardPile) Count() int {
	return len(d.cards)
}

func (d *discardPile) Empty() []ports.Card {
	c := []ports.Card{}
	copy(c, d.cards)
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
