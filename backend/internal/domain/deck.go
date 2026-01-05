package domain

import (
	"math/rand"
)

type Deck struct {
	Cards []iCard
}

func (d *Deck) DrawCard() (iCard, bool) {
	if len(d.Cards) == 0 {
		return nil, false
	}
	c := d.Cards[0]
	d.Cards = d.Cards[1:]
	return c, true
}

func NewDeck(cards []iCard) Deck {
	return Deck{Cards: cards}
}

func (d *Deck) Replenish(discardPile []iCard) {
	d.Cards = shuffle(discardPile)
}

func shuffle(cards []iCard) []iCard {
	if len(cards) == 0 {
		return cards
	}

	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	return cards
}
