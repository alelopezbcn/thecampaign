package ports

type Field interface {
	ShowCards() []Card
	GetCard(cardID string) (Card, bool)
	AddCards(cards ...Card)
	RemoveCard(card Card) bool
}
