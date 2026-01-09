package ports

type Castle interface {
	Construct(card Card) error
	IsConstructed() bool
	Value() int
	ResourceCards() int
	RemoveGold(position int) (Resource, error)
	String() string
}
