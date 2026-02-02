package ports

type Hand interface {
	ShowCards() []Card
	GetCard(cardID string) (Card, bool)
	AddCards(cards ...Card) error
	RemoveCard(card Card) bool
	CanAddCards(count int) bool
	Count() int
}
