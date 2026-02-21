package board

import (
	"errors"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
)

type Deck interface {
	Deal(players []Player)
	DrawCards(count int, discardPile DiscardPile) ([]cards.Card, error)
	Reveal(n int) []cards.Card
	Count() int
}

type deck struct {
	cards  []cards.Card
	dealer cards.Dealer
}

func newDeck(d cards.Dealer) *deck {
	return &deck{dealer: d}
}

func (d *deck) Reveal(n int) []cards.Card {
	if n > len(d.cards) {
		n = len(d.cards)
	}
	return d.cards[:n]
}

func (d *deck) Count() int {
	return len(d.cards)
}

func (d *deck) Deal(players []Player) {
	warriorCards := shuffle(d.dealer.WarriorsCards(len(players)))

	// Each player gets 3 cards.Warrior cards
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

func (d *deck) DrawCards(count int, discardPile DiscardPile) (
	cardsDrew []cards.Card, err error) {
	cardsDrew = make([]cards.Card, 0, count)
	for i := 0; i < count; i++ {
		c, ok := d.drawCard()
		if !ok {
			d.shuffleDiscardPileIntoDeck(discardPile)

			c, ok = d.drawCard()
			if !ok {
				return nil, errors.New("no cards left to draw")
			}
		}

		cardsDrew = append(cardsDrew, c)
	}

	return cardsDrew, nil
}

func (d *deck) drawCard() (cards.Card, bool) {
	if len(d.cards) == 0 {
		return nil, false
	}
	c := d.cards[0]
	d.cards = d.cards[1:]
	return c, true
}

func (d *deck) shuffleDiscardPileIntoDeck(discardPile DiscardPile) {
	d.replenish(discardPile.Empty())
}

func (d *deck) replenish(discardPile []cards.Card) {
	d.cards = shuffle(discardPile)
}

func shuffle(cards []cards.Card) []cards.Card {
	if len(cards) == 0 {
		return cards
	}

	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	return cards
}
