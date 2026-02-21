package board

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type Player interface {
	Name() string
	Idx() int
	TakeCards(cards ...cards.Card) bool
	MoveCardToField(cardID string) error
	RemoveFromHand(cardIDs ...string) ([]cards.Card, error)
	Hand() Hand
	Field() Field
	CanTakeCards(count int) bool
	CardsInHand() int
	GetCardFromHand(cardID string) (cards.Card, bool)
	GetCardFromField(cardID string) (cards.Card, bool)
	Attack(target cards.Attackable, weapon cards.Weapon) error
	UseSpecialPower(usedBy cards.Warrior, usedOn cards.Warrior,
		specialPowerCard cards.SpecialPower) error
	CardStolenFromHand(position int) (cards.Card, error)
	Construct(cardID string) error
	CanAttack() bool
	CanBuy() bool
	CanBuyWith(resource cards.Resource) bool
	CanConstruct() bool
	HasThief() bool
	HasSpy() bool
	HasCatapult() bool
	HasHarpoon() bool
	HasWarriorsInHand() bool
	CanTradeCards() bool
	Thief() cards.Thief
	Spy() cards.Spy
	Catapult() cards.Catapult
	Harpoon() cards.Harpoon
	Castle() Castle
}

type player struct {
	name                           string
	idx                            int
	hand                           Hand
	field                          Field
	castle                         Castle
	cardMovedToPileObserver        cards.CardMovedToPileObserver
	warriorMovedToCemeteryObserver WarriorMovedToCemeteryObserver
}

func NewPlayer(name string,
	idx int,
	cardMovedToPileObserver cards.CardMovedToPileObserver,
	warriorMovedToCemeteryObserver WarriorMovedToCemeteryObserver,
	castleCompletionObserver CastleCompletionObserver,
	fieldWithoutWarriorsObserver FieldWithoutWarriorsObserver,
	castleResourcesToWin int,
) *player {
	p := &player{
		name:                           name,
		idx:                            idx,
		hand:                           NewHand(),
		field:                          NewField(name, fieldWithoutWarriorsObserver),
		cardMovedToPileObserver:        cardMovedToPileObserver,
		warriorMovedToCemeteryObserver: warriorMovedToCemeteryObserver,
	}
	p.castle = NewCastle(castleResourcesToWin, p, castleCompletionObserver)

	return p
}

func (p *player) Name() string {
	return p.name
}

func (p *player) Idx() int {
	return p.idx
}

func (p *player) CanTakeCards(count int) bool {
	return p.hand.CanAddCards(count)
}

func (p *player) TakeCards(cardsTaken ...cards.Card) bool {
	if !p.hand.CanAddCards(len(cardsTaken)) {
		return false
	}

	for _, c := range cardsTaken {
		c.AddCardMovedToPileObserver(p)
		if w, ok := c.(cards.Warrior); ok {
			w.AddWarriorDeadObserver(p)
		}
	}
	_ = p.hand.AddCards(cardsTaken...)

	return true
}

func (p *player) RemoveFromHand(cardIDs ...string) ([]cards.Card, error) {
	cards := make([]cards.Card, 0, len(cardIDs))

	for _, cardID := range cardIDs {
		c, ok := p.GetCardFromHand(cardID)
		if !ok {
			return nil, fmt.Errorf("card with ID %s not found in hand", cardID)
		}

		cards = append(cards, c)
	}

	for _, c := range cards {
		p.hand.RemoveCard(c)
	}

	return cards, nil
}

func (p *player) CardsInHand() int {
	return len(p.hand.ShowCards())
}

func (p *player) Hand() Hand {
	return p.hand
}

func (p *player) Field() Field {
	return p.field
}

func (p *player) CardStolenFromHand(position int) (cards.Card, error) {
	cardsShown := p.hand.ShowCards()
	if position < 1 || position > len(cardsShown) {
		return nil, fmt.Errorf("invalid position %d for stealing cardBase", position)
	}

	// Create a copy of c.resources and shuffle it
	copied := make([]cards.Card, len(cardsShown))
	copy(copied, cardsShown)
	// Shuffle copied slice
	for i := range copied {
		j := i + rand.Intn(len(copied)-i)
		copied[i], copied[j] = copied[j], copied[i]
	}

	c := copied[position-1]
	p.hand.RemoveCard(c)

	return c, nil
}

func (p *player) GetCardFromHand(cardID string) (cards.Card, bool) {
	return p.hand.GetCard(cardID)
}

func (p *player) GetCardFromField(cardID string) (cards.Card, bool) {
	return p.field.GetWarrior(cardID)
}

func (p *player) MoveCardToField(cardID string) error {
	c, ok := p.GetCardFromHand(cardID)
	if !ok {
		return fmt.Errorf("card with ID %s not found in hand", cardID)
	}

	w, ok := c.(cards.Warrior)
	if !ok {
		return fmt.Errorf("onlywarrior or dragon cards can be moved to field")
	}

	p.field.AddWarriors(w)
	p.hand.RemoveCard(c)

	return nil
}

func (p *player) Attack(target cards.Attackable,
	weapon cards.Weapon,
) error {
	switch weapon.Type() {
	case types.SwordWeaponType:
		if !p.Field().HasKnight() && !p.Field().HasDragon() {
			return fmt.Errorf("sword weapon cannot be used")
		}
	case types.ArrowWeaponType:
		if !p.Field().HasArcher() && !p.Field().HasDragon() {
			return fmt.Errorf("arrow weapon cannot be used")
		}
	case types.PoisonWeaponType:
		if !p.Field().HasMage() && !p.Field().HasDragon() {
			return fmt.Errorf("poison weapon cannot be used")
		}
	}

	err := target.BeAttacked(weapon)
	if err != nil {
		return fmt.Errorf("attack failed: %w", err)
	}

	p.hand.RemoveCard(weapon)

	return nil
}

func (p *player) UseSpecialPower(usedBy cards.Warrior, usedOn cards.Warrior,
	specialPowerCard cards.SpecialPower,
) error {
	err := specialPowerCard.Use(usedBy, usedOn)
	if err != nil {
		return fmt.Errorf("special power failed: %w", err)
	}

	p.hand.RemoveCard(specialPowerCard)

	return nil
}

func (p *player) CanAttack() bool {
	for _, c := range p.hand.ShowCards() {
		if w, ok := c.(cards.Weapon); ok {
			if p.field.HasDragon() {
				return true
			}

			switch w.Type() {
			case types.ArrowWeaponType:
				if p.Field().HasArcher() {
					return true
				}
			case types.PoisonWeaponType:
				if p.field.HasMage() {
					return true
				}
			case types.SwordWeaponType:
				if p.field.HasKnight() {
					return true
				}
			case types.SpecialPowerWeaponType:
				// SpecialPower can be used by Archer, Knight, or Mage
				if p.field.HasArcher() || p.field.HasKnight() || p.field.HasMage() {
					return true
				}
			}
		}
	}

	return false
}

func (p *player) CanBuy() bool {
	for _, c := range p.hand.ShowCards() {
		if r, ok := c.(cards.Resource); ok {
			if p.CanBuyWith(r) {
				return true
			}
		}
	}

	return false
}

func (p *player) CanBuyWith(resource cards.Resource) bool {
	if resource.CanConstruct() {
		return false
	}

	cardsToBuy := resource.Value() / 2
	if p.Hand().Count()+cardsToBuy-1 > MaxCardsInHand {
		return false
	}

	return true
}

func (p *player) CanConstruct() bool {
	for _, c := range p.hand.ShowCards() {
		if r, ok := c.(cards.Resource); ok {
			// If castle is already constructed, any resource can be added
			if p.castle.IsConstructed() || r.CanConstruct() {
				return true
			}
		}
		if w, ok := c.(cards.Weapon); ok {
			if !p.castle.IsConstructed() && w.CanConstruct() {
				return true
			}
		}
	}

	return false
}

func (p *player) HasThief() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(cards.Thief); ok {
			return true
		}
	}
	return false
}

func (p *player) HasCatapult() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(cards.Catapult); ok {
			return true
		}
	}
	return false
}

func HasCardTypeInHand[T any](p Player) (T, bool) {
	for _, c := range p.Hand().ShowCards() {
		if card, ok := c.(T); ok {
			return card, true
		}
	}
	
	var zero T
	return zero, false
}

func (p *player) HasHarpoon() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(cards.Harpoon); ok {
			return true
		}
	}
	return false
}

func (p *player) HasWarriorsInHand() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(cards.Warrior); ok {
			return true
		}
	}
	return false
}

func (p *player) CanTradeCards() bool {
	count := 0
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(cards.Weapon); ok {
			count++
			if count >= 3 {
				return true
			}
		}
	}
	return false
}

func (p *player) Thief() cards.Thief {
	for _, c := range p.hand.ShowCards() {
		if t, ok := c.(cards.Thief); ok {
			p.hand.RemoveCard(t)
			return t
		}
	}
	return nil
}

func (p *player) HasSpy() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(cards.Spy); ok {
			return true
		}
	}
	return false
}

func (p *player) Spy() cards.Spy {
	for _, c := range p.hand.ShowCards() {
		if s, ok := c.(cards.Spy); ok {
			p.hand.RemoveCard(s)
			return s
		}
	}
	return nil
}

func (p *player) Catapult() cards.Catapult {
	for _, c := range p.hand.ShowCards() {
		if t, ok := c.(cards.Catapult); ok {
			p.hand.RemoveCard(t)
			return t
		}
	}
	return nil
}

func (p *player) Harpoon() cards.Harpoon {
	for _, c := range p.hand.ShowCards() {
		if h, ok := c.(cards.Harpoon); ok {
			p.hand.RemoveCard(h)
			return h
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

	p.hand.RemoveCard(resourceCard)

	return nil
}

func (p *player) OnCardMovedToPile(card cards.Card) {
	p.cardMovedToPileObserver.OnCardMovedToPile(card)
}

func (p *player) OnWarriorDead(warrior cards.Warrior) {
	if !p.field.RemoveWarrior(warrior) {
		fmt.Println("warrior not found in player field")
	}
	p.warriorMovedToCemeteryObserver.OnWarriorMovedToCemetery(warrior)
}
