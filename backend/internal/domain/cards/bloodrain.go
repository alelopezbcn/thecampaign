package cards

import (
	"errors"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

const (
	bloodRainDamage = 4
)

type BloodRain interface {
	Card
	Weapon
	Attack(targets []Warrior) error
}

type bloodRain struct {
	*cardBase
	*weaponBase
}

func NewBloodRain(id string) *bloodRain {
	return &bloodRain{
		cardBase:   newCardBase(id, "BloodRain"),
		weaponBase: newWeaponBase(bloodRainDamage, types.BloodRainWeaponType),
	}
}

func (b *bloodRain) Attack(targets []Warrior) error {
	if len(targets) == 0 {
		return errors.New("targets cannot be empty")
	}

	for _, target := range targets {
		target.ReceiveDamage(b, 1)
	}

	return nil
}
