package ports

type Card interface {
	GetID() string
	AssignedToPlayer(player Player)
	String() string
	GetCardToBeDiscardedObserver() CardToBeDiscardedObserver
	AddMessageObserver(o MessageObserver)
	GetMessageObserver() MessageObserver
}

type Attackable interface {
	Health() int
	ReceiveDamage(weapon Weapon, multiplier int) (isDefeated bool)
	AttackedBy() []Weapon
}

type Catapult interface {
	Card
	Attack(castle Castle, position int) (Resource, error)
}

type Resource interface {
	Card
	Value() int
}

type SpecialPower interface {
	Card
	Attackable
	Use(usedBy Warrior, target Warrior) error
	Destroyed()
}
type Spy interface {
	Card
}
type Thief interface {
	Card
}

type Warrior interface {
	Card
	Attackable
	Attack(target Attackable, weapon Weapon) error
	ProtectedBy(powerCard SpecialPower)
	Heal()
	InstantKill()
	AddWarriorDeadObserver(o WarriorDeadObserver)
}

type Weapon interface {
	Card
	DamageAmount() int
}
