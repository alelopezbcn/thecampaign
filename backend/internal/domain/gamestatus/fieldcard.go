package gamestatus

import (
	"github.com/alelopezbcn/thecampaign/internal/domain/cards"
	"github.com/alelopezbcn/thecampaign/internal/domain/types"
)

type FieldCard struct {
	Card
	AttackedBy  []Card `json:"attacked_by,omitempty"`
	ProtectedBy *Card  `json:"protected_by,omitempty"`
}

func NewFieldCard(warrior cards.Warrior) FieldCard {

	var aCardType CardType
	switch warrior.Type() {
	case types.KnightWarriorType:
		aCardType = CardTypeKnight
	case types.ArcherWarriorType:
		aCardType = CardTypeArcher
	case types.MageWarriorType:
		aCardType = CardTypeMage
	case types.DragonWarriorType:
		aCardType = CardTypeDragon
	case types.MercenaryWarriorType:
		aCardType = CardTypeMercenary
	}

	var attackedByCards []Card
	for _, attacker := range warrior.AttackedBy() {
		var weaponCardType CardType
		switch attacker.Type() {
		case types.SwordWeaponType:
			weaponCardType = CardTypeSword
		case types.ArrowWeaponType:
			weaponCardType = CardTypeArrow
		case types.PoisonWeaponType:
			weaponCardType = CardTypePoison
		}
		attackedByCards = append(attackedByCards, newCard(attacker.GetID(),
			weaponCardType, attacker.DamageAmount()))
	}

	var protectedByCard *Card
	if ok, shield := warrior.IsProtected(); ok {
		c := newCard(shield.GetID(), CardTypeSpecialPower, shield.Health())
		protectedByCard = &c
	}

	return FieldCard{
		Card:        newCard(warrior.GetID(), aCardType, warrior.Health()),
		AttackedBy:  attackedByCards,
		ProtectedBy: protectedByCard,
	}
}
