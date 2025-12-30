package domain

type Player struct {
	ID           string
	Name         string
	TurnPosition int
	Hand         []Card
	Field        []Card
	Castle       []Card
}
