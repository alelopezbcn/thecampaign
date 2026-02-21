package board

import "github.com/alelopezbcn/thecampaign/internal/domain/cards"

type board struct {
	deck        Deck
	discardPile DiscardPile
	cemetery    Cemetery
	players     []Player
}

func New(dealer cards.Dealer, players []Player) *board {
	return &board{
		deck:        newDeck(dealer),
		discardPile: newDiscardPile(),
		cemetery:    newCemetery(),
		players:     players,
	}
}

func (b *board) Deck() Deck {
	return b.deck
}

func (b *board) DiscardPile() DiscardPile {
	return b.discardPile
}

func (b *board) Cemetery() Cemetery {
	return b.cemetery
}

func (b *board) Players() []Player {
	return b.players
}
