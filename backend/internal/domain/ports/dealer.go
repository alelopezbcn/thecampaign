package ports

type Dealer interface {
	WarriorsCards() (warriors []Card)
	OtherCards() (other []Card)
}
