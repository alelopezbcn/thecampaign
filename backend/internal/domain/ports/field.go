package ports

type Field interface {
	Warriors() []Warrior
	GetWarrior(cardID string) (Warrior, bool)
	AddWarriors(cards ...Warrior)
	RemoveWarrior(card Warrior) bool
	HasArcher() bool
	HasMage() bool
	HasKnight() bool
	HasDragon() bool
	AttackableIDs() []string
}
