package domain

import "fmt"

type Player struct {
	Name   string
	hand   []iCard
	field  []iCard
	castle *Castle
}

func NewPlayer(name string) *Player {
	return &Player{
		Name:  name,
		hand:  []iCard{},
		field: []iCard{},
	}
}

func (p *Player) takeCards(cards ...iCard) {
	p.hand = append(p.hand, cards...)
}

func (p *Player) ShowHand() []iCard {
	return p.hand
}

func (p *Player) ShowField() []iCard {
	return p.field
}

func (p *Player) ShowCastle() *Castle {
	return p.castle
}

func (p *Player) GetCardFromHand(cardID string) (iCard, bool) {
	for _, c := range p.hand {
		if c.GetID() == cardID {
			return c, true
		}
	}

	return nil, false
}

func (p *Player) GetCardFromField(cardID string) (iCard, bool) {
	for _, c := range p.field {
		if c.GetID() == cardID {
			return c, true
		}
	}

	return nil, false
}

func (p *Player) moveWarriorToField(cardID string) bool {
	for i, c := range p.hand {
		if c.GetID() == cardID && c.IsWarrior() {
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
		if c.GetID() == cardID {
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

func (p *Player) Attack(warriorCard iCard, targetCard iCard, weaponCard iCard) {
	// check if player has Dragon, then can attack

	// attack with warriorCard on targetCard using weaponCard
	if err := warriorCard.Attack(targetCard, weaponCard); err != nil {
		return fmt.Errorf("attack failed: %w", err)
	}

}
