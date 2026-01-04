package domain

import (
	"fmt"
	"strings"
)

const (
	WarriorHealth     = 10
	DragonHealth      = 20
	MaxHandSize       = 7
	SpecialMoveHealth = 10
)

type Card struct {
	ID         string
	Name       string
	Type       CardType
	Value      int
	AffectedBy []Card
}

func WarriorsCards() (warriors []Card) {
	warriors = make([]Card, 0, 15)
	for i := 1; i < 6; i++ {
		warriors = append(warriors, Card{ID: fmt.Sprintf("k%d", i), Name: "Knight", Type: CardTypeKnight, Value: WarriorHealth})
		warriors = append(warriors, Card{ID: fmt.Sprintf("a%d", i), Name: "Archer", Type: CardTypeArcher, Value: WarriorHealth})
		warriors = append(warriors, Card{ID: fmt.Sprintf("m%d", i), Name: "Mage", Type: CardTypeMage, Value: WarriorHealth})
	}

	return warriors
}

func OtherButWarriorsCards() (cards []Card) {
	cards = []Card{
		{ID: "d", Name: "Dragon", Type: CardTypeDragon, Value: DragonHealth},
		{ID: "s1", Name: "Special Move", Type: CardTypeSpecialMove, Value: SpecialMoveHealth},
		{ID: "s2", Name: "Special Move", Type: CardTypeSpecialMove, Value: SpecialMoveHealth},
		{ID: "s3", Name: "Special Move", Type: CardTypeSpecialMove, Value: SpecialMoveHealth},
		{ID: "s4", Name: "Special Move", Type: CardTypeSpecialMove, Value: SpecialMoveHealth},
		{ID: "sp", Name: "Spy", Type: CardTypeSpy, Value: 0},
		{ID: "t", Name: "Thief", Type: CardTypeThief, Value: 0},
		{ID: "c", Name: "Catapult", Type: CardTypeCatapult, Value: 0},
	}

	for i := 1; i < 10; i++ {
		cards = append(cards, Card{ID: fmt.Sprintf("goA%d", i), Name: "Gold Coin", Type: CardTypeMoney, Value: i})
		cards = append(cards, Card{ID: fmt.Sprintf("goB%d", i), Name: "Gold Coin", Type: CardTypeMoney, Value: i})
		cards = append(cards, Card{ID: fmt.Sprintf("swA%d", i), Name: "Sword", Type: CardTypeSword, Value: i})
		cards = append(cards, Card{ID: fmt.Sprintf("swB%d", i), Name: "Sword", Type: CardTypeSword, Value: i})
		cards = append(cards, Card{ID: fmt.Sprintf("arA%d", i), Name: "Arrow", Type: CardTypeArrow, Value: i})
		cards = append(cards, Card{ID: fmt.Sprintf("arB%d", i), Name: "Arrow", Type: CardTypeArrow, Value: i})
		cards = append(cards, Card{ID: fmt.Sprintf("poA%d", i), Name: "Possion", Type: CardTypePossion, Value: i})
		cards = append(cards, Card{ID: fmt.Sprintf("poB%d", i), Name: "Possion", Type: CardTypePossion, Value: i})
	}

	return shuffle(cards)
}

func (c *Card) IsWarrior() bool {
	return c.Type.IsWarrior
}

func (c *Card) IsResource() bool {
	return c.Type.CanBuy
}

func (c *Card) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s (%s)", c.Name, c.ID))
	if c.Value > 0 {
		sb.WriteString(fmt.Sprintf(" - Value: %d", c.Value))
	}
	if c.AffectedBy != nil && len(c.AffectedBy) > 0 {
		for _, card := range c.AffectedBy {
			sb.WriteString(fmt.Sprintf("\n  * %s", card.String()))
		}
	}
	return sb.String()
}

type CardType struct {
	IsWarrior bool
	IsWeapon  bool
	CanAttack bool
	CanBuy    bool
	CanSteal  bool
	CanSpy    bool
	IsSpecial bool
}

var (
	CardTypeKnight      = CardType{true, false, true, false, false, false, false}
	CardTypeArcher      = CardType{true, false, true, false, false, false, false}
	CardTypeMage        = CardType{true, false, true, false, false, false, false}
	CardTypeSword       = CardType{false, true, true, false, false, false, false}
	CardTypeArrow       = CardType{false, true, true, false, false, false, false}
	CardTypePossion     = CardType{false, true, true, false, false, false, false}
	CardTypeDragon      = CardType{false, false, true, false, false, false, false}
	CardTypeSpecialMove = CardType{false, false, true, false, false, false, true}
	CardTypeSpy         = CardType{false, false, false, false, false, true, false}
	CardTypeThief       = CardType{false, false, false, false, true, false, false}
	CardTypeMoney       = CardType{false, false, false, true, false, false, false}
	CardTypeCatapult    = CardType{false, false, true, false, false, false, false}
)
