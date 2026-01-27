package ports

type Player interface {
	Name() string
	TakeCards(cards ...Card) bool
	MoveCardToField(cardID string) error
	GiveCards(cardIDs ...string) ([]Card, error)
	Hand() Hand
	Field() Field
	CanTakeCards(count int) bool
	CardsInHand() int
	GetCardFromHand(cardID string) (Card, bool)
	GetCardFromField(cardID string) (Card, bool)
	Attack(targetCard Card, weaponCard Card) error
	UseSpecialPower(warriorCard Card, targetCard Card, specialPowerCard Card) error
	CardStolenFromHand(position int) (Card, error)
	Construct(cardID string) error
	CanAttack() bool
	CanBuy() bool
	CanConstruct() bool
	Thief() Thief
	Spy() Spy
	Catapult() Catapult
	Castle() Castle
}
