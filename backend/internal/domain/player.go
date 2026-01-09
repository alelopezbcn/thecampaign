package domain

import (
	"errors"
	"fmt"
	"math/rand"
)

type Player interface {
	Name() string
	TakeCards(cards ...Card) bool
	MoveCardToField(cardID string) error
	GiveCards(cardIDs ...string) ([]Card, error)
	ShowHand() []Card
	ShowField() []Card
	CanTakeCards(count int) bool
	CardsInHand() int
	GetCardFromHand(cardID string) (Card, bool)
	GetCardFromField(cardID string) (Card, bool)
	Attack(warriorCard Card, targetCard Card, weaponCard Card) error
	UseSpecialPower(warriorCard Card, targetCard Card, specialPowerCard Card) error
	CardStolenFromHand(position int) (Card, error)
	Construct(cardID string) error
	Thief() Thief
	Spy() Spy
	Catapult() Catapult
	Castle() Castle
}

type player struct {
	name                           string
	hand                           Hand
	field                          Field
	castle                         Castle
	cardMovedToPileObserver        CardMovedToPileObserver
	warriorMovedToCemeteryObserver WarriorMovedToCemeteryObserver
}

func NewPlayer(name string,
	cardMovedToPileObserver CardMovedToPileObserver,
	warriorMovedToCemeteryObserver WarriorMovedToCemeteryObserver,
	castleCompletionObserver CastleCompletionObserver,
	fieldWithoutWarriorsObserver FieldWithoutWarriorsObserver,
) Player {
	p := &player{
		name:                           name,
		hand:                           newHand(),
		field:                          newField(fieldWithoutWarriorsObserver),
		cardMovedToPileObserver:        cardMovedToPileObserver,
		warriorMovedToCemeteryObserver: warriorMovedToCemeteryObserver,
	}
	p.castle = newCastle(p, castleCompletionObserver)

	return p
}

func (p *player) Name() string {
	return p.name
}

func (p *player) CanTakeCards(count int) bool {
	return p.hand.CanAddCards(count)
}

func (p *player) TakeCards(cards ...Card) bool {
	if !p.hand.CanAddCards(len(cards)) {
		return false
	}

	for _, c := range cards {
		c.AssignedToPlayer(p)
		if w, ok := c.(Warrior); ok {
			w.AddWarriorDeadObserver(p)
		}
	}
	_ = p.hand.AddCards(cards...)

	return true
}

func (p *player) GiveCards(cardIDs ...string) ([]Card, error) {
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

func (p *player) CardsInHand() int {
	return len(p.hand.ShowCards())
}

func (p *player) ShowHand() []Card {
	return p.hand.ShowCards()
}

func (p *player) ShowField() []Card {
	return p.field.ShowCards()
}

func (p *player) CardStolenFromHand(position int) (Card, error) {
	cards := p.hand.ShowCards()
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

func (p *player) GetCardFromHand(cardID string) (Card, bool) {
	return p.hand.GetCard(cardID)
}

func (p *player) GetCardFromField(cardID string) (Card, bool) {
	return p.field.GetCard(cardID)
}

func (p *player) MoveCardToField(cardID string) error {
	c, ok := p.GetCardFromHand(cardID)
	if !ok {
		return fmt.Errorf("cardBase with ID %s not found in hand", cardID)
	}

	if _, ok = c.(Warrior); !ok {
		return fmt.Errorf("only Warrior or dragon cards can be moved to field")
	}

	p.field.AddCards(c)
	p.removeCardFromHand(c)

	return nil
}

func (p *player) Attack(warriorCard Card, targetCard Card,
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

func (p *player) UseSpecialPower(usedBy Card, usedOn Card,
	specialPowerCard Card) error {

	s, ok := specialPowerCard.(SpecialPower)
	if !ok {
		return fmt.Errorf("the card is not a special power")
	}
	w, ok := usedBy.(Warrior)
	if !ok {
		return fmt.Errorf("the attacking card is not a warrior")
	}
	t, ok := usedOn.(Warrior)
	if !ok {
		return fmt.Errorf("the target card is not a warrior")
	}

	err := s.Use(w, t)
	if err != nil {
		return fmt.Errorf("special power failed: %w", err)
	}

	return nil
}

func (p *player) removeCardFromHand(card Card) bool {
	return p.hand.RemoveCard(card)
}

func (p *player) Thief() Thief {
	for _, c := range p.hand.ShowCards() {
		if t, ok := c.(Thief); ok {
			return t
		}
	}
	return nil
}

func (p *player) Spy() Spy {
	for _, c := range p.hand.ShowCards() {
		if s, ok := c.(Spy); ok {
			return s
		}
	}
	return nil
}

func (p *player) Catapult() Catapult {
	for _, c := range p.hand.ShowCards() {
		if t, ok := c.(Catapult); ok {
			return t
		}
	}
	return nil
}

func (p *player) Castle() Castle {
	return p.castle
}

func (p *player) Construct(cardID string) error {
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

func (p *player) OnCardToBeDiscarded(card Card) {
	if !p.removeCardFromHand(card) || !p.field.RemoveCard(card) {
		panic("card not found in player")
	}

	p.cardMovedToPileObserver.OnCardMovedToPile(card)
}

func (p *player) OnWarriorDead(warrior Warrior) {
	if !p.field.RemoveCard(warrior) {
		panic("warrior not found in player field")
	}
	p.warriorMovedToCemeteryObserver.OnWarriorMovedToCemetery(warrior)
}
