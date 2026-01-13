package domain

import (
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type deck struct {
	cards []ports.Card
}

func NewDeck(cards []ports.Card) ports.Deck {
	return &deck{cards: cards}
}

func (d *deck) DrawCard() (ports.Card, bool) {
	if len(d.cards) == 0 {
		return nil, false
	}
	c := d.cards[0]
	d.cards = d.cards[1:]
	return c, true
}

func (d *deck) Replenish(discardPile []ports.Card) {
	d.cards = shuffle(discardPile)
}

func (d *deck) Reveal(n int) []ports.Card {
	if n > len(d.cards) {
		n = len(d.cards)
	}
	return d.cards[:n]
}

func shuffle(cards []ports.Card) []ports.Card {
	if len(cards) == 0 {
		return cards
	}

	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	return cards
}
