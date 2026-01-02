package domain

import "fmt"

const (
	WarriorHealth     = 10
	DragonHealth      = 20
	MaxHandSize       = 7
	SpecialMoveHealth = 10
)

type Card interface{}

type card struct {
	ID         string
	Name       string
	Type       CardType
	Value      int
	AffectedBy []Card
}

func WarriorsCards() (warriors []card) {
	warriors = make([]card, 0, 15)
	for i := 1; i < 6; i++ {
		warriors = append(warriors, card{ID: fmt.Sprintf("knight%d", i), Name: "Knight", Type: CardTypeKnight, Value: WarriorHealth})
		warriors = append(warriors, card{ID: fmt.Sprintf("archer%d", i), Name: "Archer", Type: CardTypeArcher, Value: WarriorHealth})
		warriors = append(warriors, card{ID: fmt.Sprintf("mage%d", i), Name: "Mage", Type: CardTypeMage, Value: WarriorHealth})
	}

	return warriors
}

func OtherButWarriorsCards() (cards []card) {
	cards = []card{
		{ID: "dragon", Name: "Dragon", Type: CardTypeDragon, Value: DragonHealth},
		{ID: "special1", Name: "Special Move", Type: CardTypeSpecialMove, Value: SpecialMoveHealth},
		{ID: "special2", Name: "Special Move", Type: CardTypeSpecialMove, Value: SpecialMoveHealth},
		{ID: "special3", Name: "Special Move", Type: CardTypeSpecialMove, Value: SpecialMoveHealth},
		{ID: "special4", Name: "Special Move", Type: CardTypeSpecialMove, Value: SpecialMoveHealth},
		{ID: "spy", Name: "Spy", Type: CardTypeSpy, Value: 0},
		{ID: "thief", Name: "Thief", Type: CardTypeThief, Value: 0},
		{ID: "catapult", Name: "Catapult", Type: CardTypeCatapult, Value: 0},
	}

	for i := 1; i < 10; i++ {
		cards = append(cards, card{ID: fmt.Sprintf("money%d", i), Name: "Gold Coin", Type: CardTypeMoney, Value: i})
		cards = append(cards, card{ID: fmt.Sprintf("money%d", i), Name: "Gold Coin", Type: CardTypeMoney, Value: i})
		cards = append(cards, card{ID: fmt.Sprintf("sword%d", i), Name: "Sword", Type: CardTypeSword, Value: i})
		cards = append(cards, card{ID: fmt.Sprintf("sword%d", i), Name: "Sword", Type: CardTypeSword, Value: i})
		cards = append(cards, card{ID: fmt.Sprintf("arrow%d", i), Name: "Arrow", Type: CardTypeArrow, Value: i})
		cards = append(cards, card{ID: fmt.Sprintf("arrow%d", i), Name: "Arrow", Type: CardTypeArrow, Value: i})
		cards = append(cards, card{ID: fmt.Sprintf("possion%d", i), Name: "Possion", Type: CardTypePossion, Value: i})
		cards = append(cards, card{ID: fmt.Sprintf("possion%d", i), Name: "Possion", Type: CardTypePossion, Value: i})
	}

	return shuffle(cards)
}

func (c *card) IsWarrior() bool {
	return c.Type.IsWarrior
}

func (c *card) IsResource() bool {
	return c.Type.CanBuy
}

func (c *card) String() string {
	return fmt.Sprintf("%s (%s) - Value: %d", c.Name, c.ID, c.Value)
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
