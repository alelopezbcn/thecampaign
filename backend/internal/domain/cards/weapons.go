package cards

import (
	"fmt"
	"strings"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type swordCard struct {
	cardBase
	weaponCardBase
}

func NewSwordCard(id string, damageAmount int) ports.Weapon {
	return &swordCard{
		cardBase: cardBase{
			id:   strings.ToUpper(id),
			name: "Sword",
		},
		weaponCardBase: weaponCardBase{
			damageAmount: damageAmount,
		},
	}
}
func (s *swordCard) String() string {
	return fmt.Sprintf("%d %s (%s)", s.damageAmount, s.name, s.id)
}

type arrowCard struct {
	cardBase
	weaponCardBase
}

func NewArrowCard(id string, damageAmount int) ports.Weapon {
	return &arrowCard{
		cardBase: cardBase{
			id:   strings.ToUpper(id),
			name: "Arrow",
		},
		weaponCardBase: weaponCardBase{
			damageAmount: damageAmount,
		},
	}
}
func (a *arrowCard) String() string {
	return fmt.Sprintf("%d %s (%s)", a.damageAmount, a.name, a.id)
}

type poisonCard struct {
	cardBase
	weaponCardBase
}

func NewPoisonCard(id string, damageAmount int) ports.Weapon {
	{
		return &poisonCard{
			cardBase: cardBase{
				id:   strings.ToUpper(id),
				name: "Poison",
			},
			weaponCardBase: weaponCardBase{
				damageAmount: damageAmount,
			},
		}
	}
}
func (p *poisonCard) String() string {
	return fmt.Sprintf("%d %s (%s)", p.damageAmount, p.name, p.id)
}
