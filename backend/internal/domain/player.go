package domain

import "github.com/google/uuid"

type Player struct {
	ID     string
	Name   string
	hand   []card
	field  []card
	castle []card
}

func NewPlayer(name string) *Player {
	return &Player{
		ID:     uuid.NewString(),
		Name:   name,
		hand:   []card{},
		field:  []card{},
		castle: []card{},
	}
}

func (p *Player) takeCards(cards ...card) {
	p.hand = append(p.hand, cards...)
}

func (p *Player) ShowHand() []card {
	return p.hand
}

func (p *Player) moveWarriorToField(cardID string) bool {
	for i, c := range p.hand {
		if c.ID == cardID && c.IsWarrior() {
			// Move card from Hand to Field
			p.field = append(p.field, c)
			p.hand = append(p.hand[:i], p.hand[i+1:]...)
			return true
		}
	}
	return false
}

func (p *Player) moveResourceToCastle(cardID string) bool {
	for i, c := range p.hand {
		if c.ID == cardID && c.IsResource() {
			// Move card from Hand to Castle
			p.castle = append(p.castle, c)
			p.hand = append(p.hand[:i], p.hand[i+1:]...)
			return true
		}
	}
	return false
}
