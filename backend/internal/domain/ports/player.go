package ports

type Player interface {
	Name() string
	TakeCards(cards ...Card) bool
	MoveCardToField(cardID string) error
	GiveCards(cardIDs ...string) ([]Card, error)
	ShowHand() []Card
	ShowField() Field
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
