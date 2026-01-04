package domain

type Player struct {
	Name   string
	hand   []Card
	field  []Card
	castle *Castle
}

func NewPlayer(name string) *Player {
	return &Player{
		Name:  name,
		hand:  []Card{},
		field: []Card{},
	}
}

func (p *Player) takeCards(cards ...Card) {
	p.hand = append(p.hand, cards...)
}

func (p *Player) ShowHand() []Card {
	return p.hand
}

func (p *Player) ShowField() []Card {
	return p.field
}

func (p *Player) ShowCastle() *Castle {
	return p.castle
}

func (p *Player) moveWarriorToField(cardID string) bool {
	for i, c := range p.hand {
		if c.ID == cardID && c.IsWarrior() {
			// Move Card from Hand to Field
			p.field = append(p.field, c)
			p.hand = append(p.hand[:i], p.hand[i+1:]...)
			return true
		}
	}
	return false
}

func (p *Player) moveResourceToCastle(cardID string) bool {
	for i, c := range p.hand {
		if c.ID == cardID {
			if p.castle == nil {
				// Initialize Castle if not present
				castle, err := NewCastle(c)
				if err != nil {
					return false
				}

				p.castle = castle
				p.hand = append(p.hand[:i], p.hand[i+1:]...)
				return true
			}

			// Move Card from Hand to Castle
			if err := p.castle.AddResource(c); err != nil {
				return false
			}

			p.hand = append(p.hand[:i], p.hand[i+1:]...)
			return true
		}
	}
	return false
}
