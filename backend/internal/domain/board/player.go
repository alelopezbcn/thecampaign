package board

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

// PlayerIdentity — name and index
type PlayerIdentity interface {
	Name() string
	Idx() int
}

// PlayerHand — hand card management
type PlayerHand interface {
	Hand() Hand
	CardsInHand() int
	CanTakeCards(count int) bool
	TakeCards(cards ...cards.Card) bool
	RemoveFromHand(cardIDs ...string) ([]cards.Card, error)
	GetCardFromHand(cardID string) (cards.Card, bool)
}

// PlayerField — field card management
type PlayerField interface {
	Field() Field
	GetCardFromField(cardID string) (cards.Card, bool)
	MoveCardToField(cardID string) error
	HasWarriorsInHand() bool
}

// PlayerCastle — castle and economy
type PlayerCastle interface {
	Castle() Castle
	CanConstruct() bool
	CanBuy() bool
	CanBuyWith(resource cards.Resource) bool
}

// PlayerCombat — action eligibility
type PlayerCombat interface {
	CanAttack() bool
	CanTradeCards() bool
}

// Player composes all roles
type Player interface {
	PlayerIdentity
	PlayerHand
	PlayerField
	PlayerCastle
	PlayerCombat
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

func (p *player) CanAttack() bool {
	for _, c := range p.hand.ShowCards() {
		if w, ok := c.(cards.Weapon); ok {
			if p.field.HasWarriorType(types.DragonWarriorType) || p.field.HasWarriorType(types.MercenaryWarriorType) {
				return true
			}

			switch w.Type() {
			case types.ArrowWeaponType:
				if p.field.HasWarriorType(types.ArcherWarriorType) {
					return true
				}
			case types.PoisonWeaponType:
				if p.field.HasWarriorType(types.MageWarriorType) {
					return true
				}
			case types.SwordWeaponType:
				if p.field.HasWarriorType(types.KnightWarriorType) {
					return true
				}
			case types.SpecialPowerWeaponType:
				// SpecialPower can be used by Archer, Knight, or Mage
				if p.field.HasWarriorType(types.ArcherWarriorType) ||
					p.field.HasWarriorType(types.KnightWarriorType) ||
					p.field.HasWarriorType(types.MageWarriorType) {
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
		if w, ok := c.(cards.Weapon); ok {
			if w.Type() == types.SpecialPowerWeaponType {
				continue
			}
			count++
			if count >= 3 {
				return true
			}
		}
	}
	return false
}

func (p *player) Castle() Castle {
	return p.castle
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

func HasCardTypeInHand[T any](p PlayerHand) (T, bool) {
	for _, c := range p.Hand().ShowCards() {
		if card, ok := c.(T); ok {
			return card, true
		}
	}

	var zero T
	return zero, false
}
