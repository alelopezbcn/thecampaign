package domain

import (
	"math/rand"
)

type Deck struct {
	Cards []Card
}

func (d *Deck) DrawCard() (Card, bool) {
	if len(d.Cards) == 0 {
		return Card{}, false
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card, true
}

func NewDeck(cards []Card) Deck {
	return Deck{Cards: cards}
}

func (d *Deck) Replenish(discardPile []Card) {
	d.Cards = shuffle(discardPile)
}

func shuffle(cards []Card) []Card {
	if len(cards) == 0 {
		return cards
	}

	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	return cards
}
