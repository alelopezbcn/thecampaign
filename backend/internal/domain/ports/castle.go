package ports

type Castle interface {
	GetID() string
	Construct(card Card) error
	IsConstructed() bool
	Value() int
	ResourceCardsCount() int
	ResourceCards() []Resource
	RemoveGold(position int) (Resource, error)
	String() string
	CanBeAttacked() bool
}
