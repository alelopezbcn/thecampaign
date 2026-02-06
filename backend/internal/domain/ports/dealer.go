package ports

type Dealer interface {
	WarriorsCards(playerCount int) (warriors []Card)
	OtherCards(playerCount int) (other []Card)
}
