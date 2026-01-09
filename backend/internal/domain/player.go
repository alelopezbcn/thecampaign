package domain

import (
	"errors"
	"fmt"
	"math/rand"
)

type Player struct {
	Name                    string
	hand                    *hand
	field                   *field
	castle                  Castle
	cardMovedToPileObserver CardMovedToPileObserver
}

func NewPlayer(name string,
	cardMovedToPileObserver CardMovedToPileObserver,
	gameEndedObserver CastleCompletionObserver) *Player {
	p := &Player{
		Name:                    name,
		hand:                    newHand(),
		field:                   newField(gameEndedObserver),
		cardMovedToPileObserver: cardMovedToPileObserver,
	}
	p.castle = newCastle(p, gameEndedObserver)

	return p
}

func (p *Player) takeCards(cards ...Card) bool {
	if !p.hand.canAddCards(len(cards)) {
		return false
	}

	for _, c := range cards {
		c.SetPlayer(p)
	}
	_ = p.hand.addCards(cards...)

	return true
}

func (p *Player) giveCards(cardIDs ...string) ([]Card, error) {
	cards := make([]Card, 0, len(cardIDs))

	for _, cardID := range cardIDs {
		c, ok := p.GetCardFromHand(cardID)
		if !ok {
			return nil, fmt.Errorf("cardBase with ID %s not found in hand", cardID)
		}

		cards = append(cards, c)
	}

	for _, c := range cards {
		p.removeCardFromHand(c)
	}

	return cards, nil
}

func (p *Player) ShowHand() []Card {
	return p.hand.showCards()
}

func (p *Player) CardsInHand() int {
	return len(p.hand.showCards())
}

func (p *Player) ShowField() []Card {
	return p.field.showCards()
}

func (p *Player) Castle() Castle {
	return p.castle
}

func (p *Player) GetCardFromHand(cardID string) (Card, bool) {
	return p.hand.getCard(cardID)
}

func (p *Player) GetCardFromField(cardID string) (Card, bool) {
	return p.field.getCard(cardID)
}

func (p *Player) moveCardToField(cardID string) error {
	c, ok := p.GetCardFromHand(cardID)
	if !ok {
		return fmt.Errorf("cardBase with ID %s not found in hand", cardID)
	}

	switch c.(type) {
	case Warrior, *dragonCard:
		break
	default:
		return fmt.Errorf("only Warrior or dragon cards can be moved to field")
	}

	p.field.addCards(c)
	p.removeCardFromHand(c)

	return nil
}

func (p *Player) Attack(warriorCard Card, targetCard Card,
	weaponCard Card) error {

	warrior, ok := warriorCard.(Warrior)
	if !ok {
		return fmt.Errorf("the attacking cardBase is not a Warrior")
	}
	target, ok := targetCard.(Attackable)
	if !ok {
		return fmt.Errorf("the target cardBase cannot be attacked")
	}
	weapon, ok := weaponCard.(Weapon)
	if !ok {
		return fmt.Errorf("the Weapon cardBase is not a Weapon")
	}

	err := warrior.Attack(target, weapon)
	if err != nil {
		return fmt.Errorf("attack failed: %w", err)
	}

	return nil
}

func (p *Player) UseSpecialPower(warriorCard Card, targetCard Card,
	specialPowerCard Card) error {

	s, ok := specialPowerCard.(SpecialPower)
	if !ok {
		return fmt.Errorf("the cardBase is not a Special Power")
	}
	w, ok := warriorCard.(Warrior)
	if !ok {
		return fmt.Errorf("the attacking cardBase is not a Warrior")
	}
	t, ok := targetCard.(Attackable)
	if !ok {
		return fmt.Errorf("the target cardBase cannot be attacked")
	}

	err := s.Use(w, t)
	if err != nil {
		return fmt.Errorf("special power failed: %w", err)
	}

	return nil
}

func (p *Player) removeCardFromHand(card Card) bool {
	return p.hand.removeCard(card)
}

func (p *Player) removeCardFromField(card Card) bool {
	return p.field.removeCard(card)
}

func (p *Player) GetThief() *thiefCard {
	for _, c := range p.hand.showCards() {
		if t, ok := c.(*thiefCard); ok {
			return t
		}
	}
	return nil
}

func (p *Player) GetSpy() *spyCard {
	for _, c := range p.hand.showCards() {
		if s, ok := c.(*spyCard); ok {
			return s
		}
	}
	return nil
}

func (p *Player) GetCatapult() *catapultCard {
	for _, c := range p.hand.showCards() {
		if t, ok := c.(*catapultCard); ok {
			return t
		}
	}
	return nil
}

func (p *Player) Stolen(position int) (Card, error) {
	cards := p.hand.showCards()
	if position < 1 || position > len(cards) {
		return nil, fmt.Errorf("invalid position %d for stealing cardBase", position)
	}

	// Create a copy of c.resources and shuffle it
	copied := make([]Card, len(cards))
	copy(copied, cards)
	// Shuffle copied slice
	for i := range copied {
		j := i + rand.Intn(len(copied)-i)
		copied[i], copied[j] = copied[j], copied[i]
	}

	c := copied[position-1]
	p.removeCardFromHand(c)

	return c, nil
}

func (p *Player) HasSpy() bool {
	for _, c := range p.hand.showCards() {
		if _, ok := c.(*spyCard); ok {
			return true
		}
	}
	return false
}

func (p *Player) Construct(cardID string) error {
	resourceCard, ok := p.GetCardFromHand(cardID)
	if !ok {
		return errors.New("cardBase not in hand: " + cardID)
	}

	if err := p.castle.Construct(resourceCard); err != nil {
		return err
	}

	p.removeCardFromHand(resourceCard)

	return nil
}
