package domain

import "github.com/google/uuid"

type Player struct {
	ID     string
	Name   string
	Hand   []card
	Field  []card
	Castle []card
}

func NewPlayer(name string) *Player {
	return &Player{
		ID:     uuid.NewString(),
		Name:   name,
		Hand:   []card{},
		Field:  []card{},
		Castle: []card{},
	}
}

func (p *Player) MoveWarriorToField(cardID string) bool {
	for i, c := range p.Hand {
		if c.ID == cardID && c.IsWarrior() {
			// Move card from Hand to Field
			p.Field = append(p.Field, c)
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return true
		}
	}
	return false
}

func (p *Player) MoveCardToCastle(cardID string) bool {
	for i, c := range p.Hand {
		if c.ID == cardID && c.IsResource() {
			// Move card from Hand to Castle
			p.Castle = append(p.Castle, c)
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return true
		}
	}
	return false
}
