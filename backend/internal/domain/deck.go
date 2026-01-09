package domain

import (
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type deck struct {
	Cards []ports.Card
}

func NewDeck(cards []ports.Card) ports.Deck {
	return &deck{Cards: cards}
}

func (d *deck) DrawCard() (ports.Card, bool) {
	if len(d.Cards) == 0 {
		return nil, false
	}
	c := d.Cards[0]
	d.Cards = d.Cards[1:]
	return c, true
}

func (d *deck) Replenish(discardPile []ports.Card) {
	d.Cards = shuffle(discardPile)
}

func (d *deck) Reveal(n int) []ports.Card {
	if n > len(d.Cards) {
		n = len(d.Cards)
	}
	return d.Cards[:n]
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
