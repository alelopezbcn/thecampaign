package ports

type Player interface {
	Name() string
	Idx() int
	TakeCards(cards ...Card) bool
	MoveCardToField(cardID string) error
	GiveCards(cardIDs ...string) ([]Card, error)
	Hand() Hand
	Field() Field
	CanTakeCards(count int) bool
	CardsInHand() int
	GetCardFromHand(cardID string) (Card, bool)
	GetCardFromField(cardID string) (Card, bool)
	Attack(target Attackable, weapon Weapon) error
	UseSpecialPower(usedBy Warrior, usedOn Warrior,
		specialPowerCard SpecialPower) error
	CardStolenFromHand(position int) (Card, error)
	Construct(cardID string) error
	CanAttack() bool
	CanBuy() bool
	CanBuyWith(resource Resource) bool
	CanConstruct() bool
	HasThief() bool
	HasSpy() bool
	HasCatapult() bool
	HasWarriorsInHand() bool
	CanTradeCards() bool
	Thief() Thief
	Spy() Spy
	Catapult() Catapult
	Castle() Castle
}
