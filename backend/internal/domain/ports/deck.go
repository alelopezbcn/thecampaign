package ports

type Deck interface {
	Deal(players []Player)
	DrawCards(count int, discardPile DiscardPile) ([]Card, error)
	Reveal(n int) []Card
	Count() int
}
