package ports

type Deck interface {
	DrawCard() (Card, bool)
	Replenish(discardPile []Card)
	Reveal(n int) []Card
}
