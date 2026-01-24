package ports

type Castle interface {
	GetID() string
	Construct(card Card) error
	IsConstructed() bool
	Value() int
	ResourceCards() int
	RemoveGold(position int) (Resource, error)
	String() string
	CanBeAttacked() bool
}
