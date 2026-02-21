package cards

import (
	"errors"

	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

const (
	harpoonDamage = 20
)

type Harpoon interface {
	Card
	Weapon
	Attack(target Dragon) error
}

type harpoon struct {
	*cardBase
	*weaponBase
}

func NewHarpoon(id string) *harpoon {
	return &harpoon{
		cardBase:   newCardBase(id, "Harpoon"),
		weaponBase: newWeaponBase(harpoonDamage, types.HarpoonWeaponType),
	}
}

func (s *harpoon) Attack(target Dragon) error {
	if target == nil {
		return errors.New("target cannot be nil")
	}

	target.ReceiveDamage(s, 1)

	return nil
}
