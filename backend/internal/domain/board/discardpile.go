package board

import "github.com/alelopezbcn/thecampaign/internal/domain/cards"

type DiscardPile interface {
	Count() int
	Empty() []cards.Card
	Discard(cards.Card)
	GetLast() cards.Card
	Cards() []cards.Card
}

type discardPile struct {
	cards []cards.Card
}

func NewDiscardPile() *discardPile {
	return &discardPile{
		cards: []cards.Card{},
	}
}

func (d *discardPile) Count() int {
	return len(d.cards)
}

func (d *discardPile) Empty() []cards.Card {
	c := d.cards
	d.cards = []cards.Card{}
	return c
}

func (d *discardPile) Discard(c cards.Card) {
	d.cards = append(d.cards, c)
}

func (d *discardPile) GetLast() cards.Card {
	if d.Count() == 0 {
		return nil
	}

	return d.cards[len(d.cards)-1]
}

func (d *discardPile) Cards() []cards.Card {
	return d.cards
}
