package cards

import "github.com/alelopezbcn/thecampaign/internal/domain/ports"

type attackableCardBase struct {
	health     int
	attackedBy []ports.Weapon
}

func (a *attackableCardBase) Health() int {
	return a.health
}

func (a *attackableCardBase) AttackedBy() []ports.Weapon {
	return a.attackedBy
}
