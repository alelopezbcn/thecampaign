package cards

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type sword struct {
	*cardBase
	*weaponBase
}

func NewSword(id string, damageAmount int) ports.Sword {
	return &sword{
		cardBase:   newCardBase(id, "Sword"),
		weaponBase: newWeaponBase(damageAmount, ports.SwordWeaponType),
	}
}
func (s *sword) String() string {
	return fmt.Sprintf("%d %s (%s)", s.damageAmount, s.name, s.id)
}

type arrow struct {
	*cardBase
	*weaponBase
}

func NewArrow(id string, damageAmount int) ports.Arrow {
	return &arrow{
		cardBase:   newCardBase(id, "Arrow"),
		weaponBase: newWeaponBase(damageAmount, ports.ArrowWeaponType),
	}
}
func (a *arrow) String() string {
	return fmt.Sprintf("%d %s (%s)", a.damageAmount, a.name, a.id)
}

type poison struct {
	*cardBase
	*weaponBase
}

func NewPoison(id string, damageAmount int) ports.Poison {
	{
		return &poison{
			cardBase:   newCardBase(id, "Poison"),
			weaponBase: newWeaponBase(damageAmount, ports.PoisonWeaponType),
		}
	}
}
func (p *poison) String() string {
	return fmt.Sprintf("%d %s (%s)", p.damageAmount, p.name, p.id)
}
