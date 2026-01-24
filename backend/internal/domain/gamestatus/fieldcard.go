package gamestatus

import (
	"fmt"

	"github.com/alelopezbcn/thecampaign/internal/domain/ports"
)

type fieldCard struct {
	card
	AttackedBy  []card
	ProtectedBy card
}

func newFieldCard(warrior ports.Warrior) fieldCard {

	var aCardType cardType
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

	var attackedByCards []card
	for _, attacker := range warrior.AttackedBy() {
		var weaponCardType cardType
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

	var protectedByCard card
	if ok, protector := warrior.IsProtected(); ok {
		protectedByCard = newCard(protector.GetID(), CardTypeSpecialPower, 0)
	}

	return fieldCard{
		card:        newCard(warrior.GetID(), aCardType, warrior.Health()),
		AttackedBy:  attackedByCards,
		ProtectedBy: protectedByCard,
	}
}

func (c fieldCard) String() string {
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
