package ports

type Cemetery interface {
	Count() int
	AddCorp(Warrior)
	GetLast() Warrior
	Corps() []Warrior
}
