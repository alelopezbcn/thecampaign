package gamestatus

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type FieldCard struct {
	Card
	AttackedBy  []Card
	ProtectedBy Card
}

func newFieldCard(warrior ports.Warrior) FieldCard {

	var aCardType CardType
	switch warrior.Type() {
	case ports.KnightWarriorType:
		aCardType = CardTypeKnight
	case ports.ArcherWarriorType:
		aCardType = CardTypeArcher
	case ports.MageWarriorType:
		aCardType = CardTypeMage
	case ports.DragonWarriorType:
		aCardType = CardTypeDragon
	}

	var attackedByCards []Card
	for _, attacker := range warrior.AttackedBy() {
		var weaponCardType CardType
		switch attacker.Type() {
		case ports.SwordWeaponType:
			weaponCardType = CardTypeSword
		case ports.ArrowWeaponType:
			weaponCardType = CardTypeArrow
		case ports.PoisonWeaponType:
			weaponCardType = CardTypePoison
		}
		attackedByCards = append(attackedByCards, newCard(attacker.GetID(),
			weaponCardType, attacker.DamageAmount()))
	}

	var protectedByCard Card
	if ok, protector := warrior.IsProtected(); ok {
		protectedByCard = newCard(protector.GetID(), CardTypeSpecialPower, 0)
	}

	return FieldCard{
		Card:        newCard(warrior.GetID(), aCardType, warrior.Health()),
		AttackedBy:  attackedByCards,
		ProtectedBy: protectedByCard,
	}
}

func (c FieldCard) String() string {
	return fmt.Sprintf("%s - %s (%d)%s%s",
		c.CardID,
		c.CardType.String(),
		c.Value,
		func() string {
			if len(c.AttackedBy) > 0 {
				return fmt.Sprintf(" | AttackedBy: %v", c.AttackedBy)
			}
			return ""
		}(),
		func() string {
			if c.ProtectedBy.CardID != "" {
				return fmt.Sprintf(" | ProtectedBy: %v", c.ProtectedBy)
			}
			return ""
		}(),
	)
}
