package cards

import "github.com/alelopezbcn/thecampaign/internal/domain/ports"

type attackableBase struct {
	health     int
	attackedBy []ports.Weapon
}

func newAttackableBase(health int) *attackableBase {
	return &attackableBase{
		health:     health,
		attackedBy: []ports.Weapon{},
	}
}

func (a *attackableBase) Health() int {
	return a.health
}

func (a *attackableBase) AttackedBy() []ports.Weapon {
	return a.attackedBy
}
