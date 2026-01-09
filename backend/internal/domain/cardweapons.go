package domain

import (
	"fmt"
	"strings"
)

type Weapon interface {
	Card
	DamageAmount() int
}

type weaponCardBase struct {
	damageAmount int
}

func (s *weaponCardBase) DamageAmount() int {
	return s.damageAmount
}

type swordCard struct {
	cardBase
	weaponCardBase
}

func newSwordCard(id string, damageAmount int) Weapon {
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

func newArrowCard(id string, damageAmount int) Weapon {
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

func newPoisonCard(id string, damageAmount int) Weapon {
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
