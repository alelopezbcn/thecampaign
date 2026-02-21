package cards

type attackableBase struct {
	health     int
	attackedBy []Weapon
}

type Attackable interface {
	Health() int
	ReceiveDamage(weapon Weapon, multiplier int) (isDefeated bool)
	BeAttacked(weapon Weapon) error
	AttackedBy() []Weapon
	String() string
}

func newAttackableBase(health int) *attackableBase {
	return &attackableBase{
		health:     health,
		attackedBy: []Weapon{},
	}
}

func (a *attackableBase) Health() int {
	if a.health < 0 {
		return 0
	}

	return a.health
}

func (a *attackableBase) AttackedBy() []Weapon {
	return a.attackedBy
}
