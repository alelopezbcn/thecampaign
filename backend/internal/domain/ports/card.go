package ports

import "github.com/alelopezbcn/thecampaign/internal/domain/types"

type Card interface {
	GetID() string
	String() string
	AddCardMovedToPileObserver(observer CardMovedToPileObserver)
	GetCardMovedToPileObserver() CardMovedToPileObserver
}

type Attackable interface {
	Health() int
	ReceiveDamage(weapon Weapon, multiplier int) (isDefeated bool)
	BeAttacked(weapon Weapon) error
	AttackedBy() []Weapon
}

type Catapult interface {
	Card
	Attack(castle Castle, position int) (Resource, error)
}

type Resource interface {
	Card
	Value() int
	CanConstruct() bool
	CanBuy() bool
}

type SpecialPower interface {
	Card
	Attackable
	Weapon
	Use(usedBy Warrior, target Warrior) error
	Destroyed()
}
type Spy interface {
	Card
	CanSpy()
}
type Thief interface {
	Card
	CanSteal()
}

type Warrior interface {
	Card
	Attackable
	Protect(powerCard SpecialPower) error
	IsProtected() (bool, Card)
	Heal(powerCard SpecialPower)
	InstantKill(sp SpecialPower)
	AddWarriorDeadObserver(o WarriorDeadObserver)
	Type() types.WarriorType
	IsDamaged() bool
}

type Weapon interface {
	Card
	DamageAmount() int
	Type() types.WeaponType
	CanConstruct() bool
}

type Dragon interface {
	Warrior
}

type Knight interface {
	Warrior
}

type Archer interface {
	Warrior
}

type Mage interface {
	Warrior
}

type Sword interface {
	Weapon
}

type Arrow interface {
	Weapon
}

type Poison interface {
	Weapon
}
