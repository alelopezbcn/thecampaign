package domain

import (
	"fmt"
)

type Player struct {
	Name             string
	hand             *hand
	field            *field
	castle           *Castle
	cardUsedObserver CardUsedObserver
}

func NewPlayer(name string,
	cardUsedObserver CardUsedObserver,
	gameEndedObserver GameEndedObserver) *Player {
	return &Player{
		Name:             name,
		hand:             newHand(),
		field:            newField(gameEndedObserver),
		castle:           newCastle(gameEndedObserver),
		cardUsedObserver: cardUsedObserver,
	}
}

func (p *Player) takeCards(cards ...iCard) bool {
	if !p.hand.canAddCards(len(cards)) {
		return false
	}

	for _, c := range cards {
		c.SetPlayer(p)
	}
	_ = p.hand.addCards(cards...)

	return true
}

func (p *Player) ShowHand() []iCard {
	return p.hand.showCards()
}

func (p *Player) ShowField() []iCard {
	return p.field.showCards()
}

func (p *Player) ShowCastle() *Castle {
	return p.castle
}

func (p *Player) GetCardFromHand(cardID string) (iCard, bool) {
	return p.hand.getCard(cardID)
}

func (p *Player) GetCardFromField(cardID string) (iCard, bool) {
	return p.field.getCard(cardID)
}

func (p *Player) moveCardToField(cardID string) error {
	c, ok := p.GetCardFromHand(cardID)
	if !ok {
		return fmt.Errorf("card with ID %s not found in hand", cardID)
	}

	switch c.(type) {
	case warrior, *dragonCard:
		break
	default:
		return fmt.Errorf("only warrior or dragon cards can be moved to field")
	}

	p.field.addCards(c)
	p.removeCardFromHand(c)

	return nil
}

func (p *Player) moveResourceToCastle(cardID string) error {
	c, ok := p.GetCardFromHand(cardID)
	if !ok {
		return fmt.Errorf("card with ID %s not found in hand", cardID)
	}

	res, ok := c.(resource)
	if !ok {
		return fmt.Errorf("card with ID %s is not a resource", cardID)
	}

	if !p.castle.isConstructed {
		return fmt.Errorf("castle not constructed")
	}

	if err := p.castle.AddResource(res); err != nil {
		return err
	}

	p.removeCardFromHand(c)
	return nil
}

func (p *Player) Attack(warriorCard iCard, targetCard iCard,
	weaponCard iCard) error {

	d, ok := warriorCard.(*dragonCard)
	if ok {
		return d.Attack(targetCard, weaponCard)
	}

	a, ok := warriorCard.(attacker)
	if !ok {
		return fmt.Errorf("the attacking card cannot attack")
	}

	err := a.Attack(targetCard, weaponCard)
	if err != nil {
		return fmt.Errorf("attack failed: %w", err)
	}

	if p.cardUsedObserver != nil {
		p.cardUsedObserver.OnCardUsed(p, weaponCard)
	}

	return nil
}

func (p *Player) removeCardFromHand(card iCard) bool {
	return p.hand.removeCard(card)
}

func (p *Player) removeCardFromField(card iCard) bool {
	return p.field.removeCard(card)
}
