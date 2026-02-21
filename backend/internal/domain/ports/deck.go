package ports

type Deck interface {
	Deal(players []Player)
	DrawCard() (Card, bool)
	Replenish(discardPile []Card)
	Reveal(n int) []Card
	Count() int
}
