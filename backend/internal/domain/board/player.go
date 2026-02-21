package board

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type player struct {
	name                           string
	idx                            int
	hand                           ports.Hand
	field                          ports.Field
	castle                         ports.Castle
	cardMovedToPileObserver        ports.CardMovedToPileObserver
	warriorMovedToCemeteryObserver ports.WarriorMovedToCemeteryObserver
}

func NewPlayer(name string,
	idx int,
	cardMovedToPileObserver ports.CardMovedToPileObserver,
	warriorMovedToCemeteryObserver ports.WarriorMovedToCemeteryObserver,
	castleCompletionObserver ports.CastleCompletionObserver,
	fieldWithoutWarriorsObserver ports.FieldWithoutWarriorsObserver,
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

func (p *player) TakeCards(cards ...ports.Card) bool {
	if !p.hand.CanAddCards(len(cards)) {
		return false
	}

	for _, c := range cards {
		c.AddCardMovedToPileObserver(p)
		if w, ok := c.(ports.Warrior); ok {
			w.AddWarriorDeadObserver(p)
		}
	}
	_ = p.hand.AddCards(cards...)

	return true
}

func (p *player) GiveCards(cardIDs ...string) ([]ports.Card, error) {
	cards := make([]ports.Card, 0, len(cardIDs))

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

func (p *player) Hand() ports.Hand {
	return p.hand
}

func (p *player) Field() ports.Field {
	return p.field
}

func (p *player) CardStolenFromHand(position int) (ports.Card, error) {
	cards := p.hand.ShowCards()
	if position < 1 || position > len(cards) {
		return nil, fmt.Errorf("invalid position %d for stealing cardBase", position)
	}

	// Create a copy of c.resources and shuffle it
	copied := make([]ports.Card, len(cards))
	copy(copied, cards)
	// Shuffle copied slice
	for i := range copied {
		j := i + rand.Intn(len(copied)-i)
		copied[i], copied[j] = copied[j], copied[i]
	}

	c := copied[position-1]
	p.hand.RemoveCard(c)

	return c, nil
}

func (p *player) GetCardFromHand(cardID string) (ports.Card, bool) {
	return p.hand.GetCard(cardID)
}

func (p *player) GetCardFromField(cardID string) (ports.Card, bool) {
	return p.field.GetWarrior(cardID)
}

func (p *player) MoveCardToField(cardID string) error {
	c, ok := p.GetCardFromHand(cardID)
	if !ok {
		return fmt.Errorf("card with ID %s not found in hand", cardID)
	}

	w, ok := c.(ports.Warrior)
	if !ok {
		return fmt.Errorf("onlywarrior or dragon cards can be moved to field")
	}

	p.field.AddWarriors(w)
	p.hand.RemoveCard(c)

	return nil
}

func (p *player) Attack(target ports.Attackable,
	weapon ports.Weapon,
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

func (p *player) UseSpecialPower(usedBy ports.Warrior, usedOn ports.Warrior,
	specialPowerCard ports.SpecialPower,
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
		if w, ok := c.(ports.Weapon); ok {
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
		if r, ok := c.(ports.Resource); ok {
			if p.CanBuyWith(r) {
				return true
			}
		}
	}

	return false
}

func (p *player) CanBuyWith(resource ports.Resource) bool {
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
		if r, ok := c.(ports.Resource); ok {
			// If castle is already constructed, any resource can be added
			if p.castle.IsConstructed() || r.CanConstruct() {
				return true
			}
		}
		if w, ok := c.(ports.Weapon); ok {
			if !p.castle.IsConstructed() && w.CanConstruct() {
				return true
			}
		}
	}

	return false
}

func (p *player) HasThief() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(ports.Thief); ok {
			return true
		}
	}
	return false
}

func (p *player) HasCatapult() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(ports.Catapult); ok {
			return true
		}
	}
	return false
}

func (p *player) HasWarriorsInHand() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(ports.Warrior); ok {
			return true
		}
	}
	return false
}

func (p *player) CanTradeCards() bool {
	count := 0
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(ports.Weapon); ok {
			count++
			if count >= 3 {
				return true
			}
		}
	}
	return false
}

func (p *player) Thief() ports.Thief {
	for _, c := range p.hand.ShowCards() {
		if t, ok := c.(ports.Thief); ok {
			p.hand.RemoveCard(t)
			return t
		}
	}
	return nil
}

func (p *player) HasSpy() bool {
	for _, c := range p.hand.ShowCards() {
		if _, ok := c.(ports.Spy); ok {
			return true
		}
	}
	return false
}

func (p *player) Spy() ports.Spy {
	for _, c := range p.hand.ShowCards() {
		if s, ok := c.(ports.Spy); ok {
			p.hand.RemoveCard(s)
			return s
		}
	}
	return nil
}

func (p *player) Catapult() ports.Catapult {
	for _, c := range p.hand.ShowCards() {
		if t, ok := c.(ports.Catapult); ok {
			p.hand.RemoveCard(t)
			return t
		}
	}
	return nil
}

func (p *player) Castle() ports.Castle {
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

func (p *player) OnCardMovedToPile(card ports.Card) {
	p.cardMovedToPileObserver.OnCardMovedToPile(card)
}

func (p *player) OnWarriorDead(warrior ports.Warrior) {
	if !p.field.RemoveWarrior(warrior) {
		fmt.Println("warrior not found in player field")
	}
	p.warriorMovedToCemeteryObserver.OnWarriorMovedToCemetery(warrior)
}
