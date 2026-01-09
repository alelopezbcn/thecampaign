package domain

type Attackable interface {
	Health() int
	ReceiveDamage(weapon Weapon, multiplier int) (isDefeated bool)
	AttackedBy() []Weapon
}

type attackableCardBase struct {
	health     int
	attackedBy []Weapon
}

func (a *attackableCardBase) Health() int {
	return a.health
}

func (a *attackableCardBase) AttackedBy() []Weapon {
	return a.attackedBy
}
