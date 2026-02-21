package board

import "github.com/alelopezbcn/thecampaign/internal/domain/ports"

type board struct {
	deck        ports.Deck
	discardPile ports.DiscardPile
	cemetery    ports.Cemetery
	players     []ports.Player
}

func New(dealer ports.Dealer, players []ports.Player) *board {
	return &board{
		deck:        newDeck(dealer),
		discardPile: newDiscardPile(),
		cemetery:    newCemetery(),
		players:     players,
	}
}

func (b *board) Deck() ports.Deck {
	return b.deck
}

func (b *board) DiscardPile() ports.DiscardPile {
	return b.discardPile
}

func (b *board) Cemetery() ports.Cemetery {
	return b.cemetery
}

func (b *board) Players() []ports.Player {
	return b.players
}
