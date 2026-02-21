package ports

type DiscardPile interface {
	Count() int
	Empty() []Card
	Discard(Card)
	GetLast() Card
	Cards() []Card
}
