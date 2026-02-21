package board

import (
	"errors"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type deck struct {
	cards  []ports.Card
	dealer ports.Dealer
}

func NewDeck(d ports.Dealer) *deck {
	return &deck{dealer: d}
}

func (d *deck) Reveal(n int) []ports.Card {
	if n > len(d.cards) {
		n = len(d.cards)
	}
	return d.cards[:n]
}

func (d *deck) Count() int {
	return len(d.cards)
}

func (d *deck) Deal(players []ports.Player) {
	warriorCards := shuffle(d.dealer.WarriorsCards(len(players)))

	// Each player gets 3 Warrior cards
	warriorsIdx := 0
	for _, p := range players {
		p.TakeCards(warriorCards[warriorsIdx : warriorsIdx+3]...)
		warriorsIdx += 3
	}

	deckCards := append(warriorCards[warriorsIdx:],
		d.dealer.OtherCards(len(players))...)

	deckCards = shuffle(deckCards)
	otherIdx := 0
	for _, p := range players {
		p.TakeCards(deckCards[otherIdx : otherIdx+4]...)
		otherIdx += 4
	}

	deckCards = deckCards[otherIdx:]
	d.cards = deckCards
}

func (d *deck) DrawCards(count int, discardPile ports.DiscardPile) (
	cards []ports.Card, err error) {
	cards = make([]ports.Card, 0, count)
	for i := 0; i < count; i++ {
		c, ok := d.drawCard()
		if !ok {
			d.shuffleDiscardPileIntoDeck(discardPile)

			c, ok = d.drawCard()
			if !ok {
				return nil, errors.New("no cards left to draw")
			}
		}

		cards = append(cards, c)
	}

	return cards, nil
}

func (d *deck) drawCard() (ports.Card, bool) {
	if len(d.cards) == 0 {
		return nil, false
	}
	c := d.cards[0]
	d.cards = d.cards[1:]
	return c, true
}

func (d *deck) shuffleDiscardPileIntoDeck(discardPile ports.DiscardPile) {
	d.replenish(discardPile.Empty())
}

func (d *deck) replenish(discardPile []ports.Card) {
	d.cards = shuffle(discardPile)
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
