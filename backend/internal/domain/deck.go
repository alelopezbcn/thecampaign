package domain

import (
	"math/rand"
)

type Deck struct {
	Cards []card
}

func (d *Deck) DrawCard() (card, bool) {
	if len(d.Cards) == 0 {
		return card{}, false
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card, true
}

func newDeck(cards []card) Deck {
	return Deck{Cards: cards}
}

func (d *Deck) Replenish(discardPile []card) {
	d.Cards = shuffle(discardPile)
}

func shuffle(cards []card) []card {
	if len(cards) == 0 {
		return cards
	}

	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	return cards
}
