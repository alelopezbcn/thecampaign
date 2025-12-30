package domain

type Card struct {
	ID         string
	Name       string
	Type       CardType
	Value      int
	AffectedBy []Card
}

type CardType struct {
	IsWarrior bool
	CanAttack bool
	CanBuy    bool
	CanSteal  bool
	CanSpy    bool
	IsSpecial bool
}

var (
	CardTypeKnight      = CardType{true, true, false, false, false, false}
	CardTypeArcher      = CardType{true, true, false, false, false, false}
	CardTypeMage        = CardType{true, true, false, false, false, false}
	CardTypeDragon      = CardType{false, true, false, false, false, false}
	CardTypeSpecialMove = CardType{false, true, false, false, false, true}
	CardTypeSpy         = CardType{false, false, false, false, true, false}
	CardTypeThief       = CardType{false, false, false, true, false, false}
	CardTypeMoney       = CardType{false, false, true, false, false, false}
)
